import { readFile } from "fs/promises";
import { parse } from "@babel/parser";
import traverseModule from "@babel/traverse";
import { parse as parseVue } from "@vue/compiler-sfc";
import { baseParse as parseTemplate } from "@vue/compiler-dom";
import type { NodePath } from "@babel/traverse";

const traverse = traverseModule.default;

function extractKeysFromCode(code: string, i18nFns: string[]): string[] {
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
    CallExpression(path:NodePath) {
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

function extractKeysFromVueTemplate(
  template: string,
  i18nFns: string[]
): string[] {
  const keys = new Set<string>();
  const ast = parseTemplate(template);

  function walk(node: any) {
    if (!node || typeof node !== "object") return;

    // handle {{ t('xxx') }}
    if (node.type === 5 && node.content?.content) {
      const expr = node.content.content;

      try {
        const exprAst = parse(expr, {
          sourceType: "unambiguous",
          plugins: ["typescript"],
        });

        traverse(exprAst, {
          CallExpression(path:NodePath) {
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

    // handle child nodes
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
  let config;
  try {
    const imported = await import("../i18n.config.ts");
    config = imported.default;
  } catch (err) {
    throw new Error(`load i18n.config.ts failed: ${(err as Error).message}`);
  }

  const i18nFns = config.i18nFns || ["t", "$t"];

  const allContent = await readFile(file, "utf-8");
  const keys = new Set<string>();

  if (/\.(js|ts|jsx|tsx)$/.test(file)) {
    extractKeysFromCode(allContent, i18nFns).forEach((k) => keys.add(k));
  } else if (/\.vue$/.test(file)) {
    const sfc = parseVue(allContent);
    const script = sfc.descriptor.script?.content || "";
    const scriptSetup = sfc.descriptor.scriptSetup?.content || "";
    const template = sfc.descriptor.template?.content || "";

    extractKeysFromCode(script, i18nFns).forEach((k) => keys.add(k));
    extractKeysFromCode(scriptSetup, i18nFns).forEach((k) => keys.add(k));
    extractKeysFromVueTemplate(template, i18nFns).forEach((k) => keys.add(k));
  }

  return keys.size ? Array.from(keys) : [];
}
