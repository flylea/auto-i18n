import { readFile } from "fs/promises";
import { parse } from "@babel/parser";
import traverse from "@babel/traverse";
import config from "../i18n.config";
import { parse as parseVue } from "@vue/compiler-sfc";
import { baseParse as parseTemplate } from "@vue/compiler-dom";

const i18nFns = config.i18nFns || ["t", "$t"];

function extractKeysFromCode(code: string): string[] {
  const keys = new Set<string>();
  let ast;

  try {
    ast = parse(code, {
      sourceType: "unambiguous",
      plugins: ["jsx", "typescript"],
    });
  } catch {
    return [];
  }

  traverse(ast, {
    CallExpression(path) {
      const callee = path.node.callee;

      if (
        (callee.type === "Identifier" && i18nFns.includes(callee.name)) ||
        (callee.type === "MemberExpression" &&
          callee.property &&
          "name" in callee.property &&
          i18nFns.includes(callee.property.name))
      ) {
        const arg = path.node.arguments[0];
        if (arg && arg.type === "StringLiteral") {
          keys.add(arg.value);
        }
      }
    },
  });

  return [...keys];
}

function extractKeysFromVueTemplate(template: string): string[] {
  const keys = new Set<string>();

  const ast = parseTemplate(template);

  function walk(node: any) {
    if (!node || typeof node !== "object") return;

    // 处理插值表达式 {{ t('xxx') }}
    if (node.type === 5 && node.content?.content) {
      const expr = node.content.content;

      try {
        const exprAst = parse(expr, {
          sourceType: "unambiguous",
          plugins: ["typescript"],
        });

        traverse(exprAst, {
          CallExpression(path) {
            const callee = path.node.callee;

            if (
              (callee.type === "Identifier" && i18nFns.includes(callee.name)) ||
              (callee.type === "MemberExpression" &&
                callee.property &&
                "name" in callee.property &&
                i18nFns.includes(callee.property.name))
            ) {
              const arg = path.node.arguments[0];
              if (arg?.type === "StringLiteral") {
                keys.add(arg.value);
              }
            }
          },
        });
      } catch {}
    }

    // 通用递归（核心）
    for (const key in node) {
      const value = node[key];

      if (value && typeof value === "object") {
        walk(value);
      }
    }
  }

  walk(ast);

  return [...keys];
}

export async function extractKeys(file: string): Promise<string[]> {
  const allContent = await readFile(file, "utf-8");
  const keys = new Set<string>();

  if (/\.(js|ts|jsx|tsx)$/.test(file)) {
    extractKeysFromCode(allContent).forEach((k) => keys.add(k));
  } else if (/\.vue$/.test(file)) {
    const sfc = parseVue(allContent);

    const script = sfc.descriptor.script?.content || "";
    const scriptSetup = sfc.descriptor.scriptSetup?.content || "";
    const template = sfc.descriptor.template?.content || "";

    extractKeysFromCode(script).forEach((k) => keys.add(k));
    extractKeysFromCode(scriptSetup).forEach((k) => keys.add(k));
    extractKeysFromVueTemplate(template).forEach((k) => keys.add(k));
  }

  return [...keys];
}
