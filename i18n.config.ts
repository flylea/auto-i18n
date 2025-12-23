export default {
  entryDirs: ["src/views", "src/components"],
  extensions: ["vue"],
  output: "ts",
  outputDir: "src/i18n/locale",
  withFileComment: false,
  withKeyFileComment: false,
  i18nFns: ["t", "$t"],
  supportChineseKey: true,
  languages: ["zh-CN", "en-US"],
};