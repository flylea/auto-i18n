package scanner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flylea/auto-i18n/internal/config"
	"github.com/flylea/auto-i18n/internal/types"
)

type Scanner struct {
	cfg      *config.Config
	forceAll bool // true=扫描所有文件, false=只扫描 git 变更文件
}

// 支持多种框架的 i18n 函数模式
var patterns = []struct {
	pattern  *regexp.Regexp
	groupIdx int
}{
	// 匹配 t('key') 或 $t('key')
	{regexp.MustCompile(`(?:^|[^\w])(t|\$t)\s*\(\s*['"]([^'"]+)['"]\s*\)`), 2},
	// 匹配 i18n.t('key')
	{regexp.MustCompile(`(?:^|[^\w])i18n\.t\s*\(\s*['"]([^'"]+)['"]\s*\)`), 1},
	// 匹配 useTranslation()['t']('key')
	{regexp.MustCompile(`useTranslation\(\)\s*\[\s*['"]t['"]\s*\]\s*\(\s*['"]([^'"]+)['"]\s*\)`), 1},
}

func New(cfg *config.Config, forceAll bool) *Scanner {
	return &Scanner{cfg: cfg, forceAll: forceAll}
}

// Scan 扫描翻译 key
// 根据 forceAll 决定是全量扫描还是只扫描 git 变更文件
func (s *Scanner) Scan() ([]string, types.KeyFileMap, error) {
	extSet := make(map[string]bool)
	for _, ext := range s.cfg.Extensions {
		extSet["."+ext] = true
	}

	// --all 标志：强制全量扫描
	if s.forceAll {
		fmt.Println("Scanning all files (--all mode)...")
		return s.scanAll(extSet)
	}

	// 默认：尝试获取 git 变更文件
	changedFiles := s.getChangedFiles()

	// 过滤出符合扩展名的变更文件
	var targetFiles []string
	if len(changedFiles) > 0 {
		for _, f := range changedFiles {
			if extSet[strings.ToLower(filepath.Ext(f))] {
				targetFiles = append(targetFiles, f)
			}
		}
	}

	// 无变更文件或 git 不可用时，回退到全量扫描
	if len(targetFiles) == 0 {
		fmt.Println("No git changes detected, scanning all files...")
		return s.scanAll(extSet)
	}

	fmt.Printf("Git detected %d changed files, scanning only changes...\n", len(targetFiles))
	return s.scanFiles(targetFiles, extSet)
}

// getChangedFiles 获取自上次提交以来的变更文件
// 相对于 HEAD 的变更（未 staged + 已 staged + 未提交的）
func (s *Scanner) getChangedFiles() []string {
	// 获取已 staged 文件 (git diff --cached --name-only)
	staged := s.gitExec("diff", "--cached", "--name-only")
	// 获取未 staged 文件 (git diff --name-only)
	unstaged := s.gitExec("diff", "--name-only")
	// 获取未跟踪的新文件 (git ls-files --others --exclude-standard)
	untracked := s.gitExec("ls-files", "--others", "--exclude-standard")

	all := append(append(staged, unstaged...), untracked...)
	return deduplicate(all)
}

// gitExec 执行 git 命令并返回文件列表
func (s *Scanner) gitExec(args ...string) []string {
	cmd := exec.Command("git", args...)
	cmd.Dir = s.workingDir()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return nil
	}

	lines := strings.Split(output, "\n")
	var files []string
	for _, f := range lines {
		f = strings.TrimSpace(f)
		if f != "" {
			files = append(files, f)
		}
	}
	return files
}

// workingDir 获取工作目录（使用运行命令时的当前目录）
func (s *Scanner) workingDir() string {
	// 获取当前工作目录（用户运行命令的位置）
	cwd, err := os.Getwd()
	if err != nil {
		if len(s.cfg.EntryDirs) > 0 {
			return s.cfg.EntryDirs[0]
		}
		return "."
	}
	return cwd
}

// deduplicate 去重
func deduplicate(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// scanAll 全量扫描
func (s *Scanner) scanAll(extSet map[string]bool) ([]string, types.KeyFileMap, error) {
	var allFiles []string

	for _, dir := range s.cfg.EntryDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			if !extSet[strings.ToLower(filepath.Ext(path))] {
				return nil
			}
			allFiles = append(allFiles, path)
			return nil
		})
		if err != nil {
			continue
		}
	}

	return s.scanFiles(allFiles, extSet)
}

// scanFiles 扫描指定文件列表
func (s *Scanner) scanFiles(files []string, extSet map[string]bool) ([]string, types.KeyFileMap, error) {
	keyFileMap := make(types.KeyFileMap)
	allKeys := make(map[string]bool)

	for _, file := range files {
		if !extSet[strings.ToLower(filepath.Ext(file))] {
			continue
		}

		keys, err := s.extractKeys(file)
		if err != nil {
			fmt.Printf("Warning: failed to process %s: %v\n", file, err)
			continue
		}

		for _, key := range keys {
			allKeys[key] = true
			if keyFileMap[key] == nil {
				keyFileMap[key] = make(map[string]bool)
			}
			keyFileMap[key][file] = true
		}
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}

	return keys, keyFileMap, nil
}

func (s *Scanner) extractKeys(file string) ([]string, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(file))
	if ext == ".vue" {
		return s.extractFromVue(string(content))
	}
	return s.extractFromCode(string(content)), nil
}

func (s *Scanner) extractFromVue(content string) ([]string, error) {
	keys := make(map[string]bool)

	// 提取 script 部分
	scriptMatch := regexp.MustCompile(`<script[^>]*>([\s\S]*?)</script>`).FindStringSubmatch(content)
	if len(scriptMatch) > 1 {
		for _, k := range s.extractFromCode(scriptMatch[1]) {
			keys[k] = true
		}
	}

	// 提取 template 部分
	templateMatch := regexp.MustCompile(`<template[^>]*>([\s\S]*?)</template>`).FindStringSubmatch(content)
	if len(templateMatch) > 1 {
		for _, k := range s.extractFromCode(templateMatch[1]) {
			keys[k] = true
		}
	}

	return mapToSlice(keys), nil
}

func (s *Scanner) extractFromCode(content string) []string {
	keys := make(map[string]bool)

	for _, p := range patterns {
		matches := p.pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > p.groupIdx {
				keys[match[p.groupIdx]] = true
			}
		}
	}

	return mapToSlice(keys)
}

func mapToSlice(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
