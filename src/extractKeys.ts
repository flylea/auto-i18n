import { readFile } from "fs/promises";
import { parse } from "@babel/parser";
import traverseModule from "@babel/traverse";
import { parse as parseVue } from "@vue/compiler-sfc";
import { baseParse as parseTemplate } from "@vue/compiler-dom";
import type { CallExpression } from "@babel/types";
import type { NodePath } from '@babel/traverse';

const traverse = traverseModule.default;

export function extractKeysFromCode(code: string, i18nFns: string[]) {
  const keys = new Set<string>();
  let ast;
  try {
    ast = parse(code, {
      sourceType: "unambiguous",
      plugins: ["typescript", "jsx"]
    });
  } catch {
    return [];
  }

  traverse(ast, {
    CallExpression(path: NodePath<CallExpression>) {
      const callee = path.node.callee;
      if (
        (callee.type === "Identifier" && i18nFns.includes(callee.name)) ||
        (callee.type === "MemberExpression" &&
          callee.property &&
          "name" in callee.property &&
          i18nFns.includes(callee.property.name))
      ) {
        const arg = path.node.arguments[0];
        if (arg?.type === "StringLiteral") keys.add(arg.value);
      }
    }
  });

  return Array.from(keys);
}

export function extractKeysFromVueTemplate(template: string, i18nFns: string[]) {
  const keys = new Set<string>();
  const ast = parseTemplate(template);

  function walk(node: any) {
    if (!node || typeof node !== "object") return;

    if (node.type === 5 && node.content?.content) {
      const expr = node.content.content;
      try {
        const exprAst = parse(expr, {
          sourceType: "unambiguous",
          plugins: ["typescript"]
        });
        traverse(exprAst, {
          CallExpression(path: NodePath<CallExpression>) {
            const callee = path.node.callee;
            if (
              (callee.type === "Identifier" && i18nFns.includes(callee.name)) ||
              (callee.type === "MemberExpression" &&
                callee.property &&
                "name" in callee.property &&
                i18nFns.includes(callee.property.name))
            ) {
              const arg = path.node.arguments[0];
              if (arg?.type === "StringLiteral") keys.add(arg.value);
            }
          }
        });
      } catch {}
    }

    Object.values(node).forEach((child) => {
      if (child && typeof child === "object") walk(child);
    });
  }

  walk(ast);
  return Array.from(keys);
}

export async function extractKeys(file: string) {
  let config: any;
  try {
    const imported = await import("../i18n.config.js");
    config = imported.default || imported;
  } catch (err: any) {
    throw new Error(`load i18n.config.ts failed: ${err.message}`);
  }

  const i18nFns: string[] = config.i18nFns || ["t", "$t"];
  const content = await readFile(file, "utf-8");
  const keys = new Set<string>();

  if (/\.(js|ts|jsx|tsx)$/.test(file)) {
    extractKeysFromCode(content, i18nFns).forEach((k) => keys.add(k));
  } else if (/\.vue$/.test(file)) {
    const sfc = parseVue(content);
    const script = sfc.descriptor.script?.content || "";
    const scriptSetup = sfc.descriptor.scriptSetup?.content || "";
    const template = sfc.descriptor.template?.content || "";

    extractKeysFromCode(script, i18nFns).forEach((k) => keys.add(k));
    extractKeysFromCode(scriptSetup, i18nFns).forEach((k) => keys.add(k));
    extractKeysFromVueTemplate(template, i18nFns).forEach((k) => keys.add(k));
  }

  return Array.from(keys);
}
