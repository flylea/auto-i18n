export default {
  "entryDirs": [
    "playground/src/views",
    "playground/src/pages",
    "playground/src/components"
  ],
  "extensions": [
    "vue",
    "tsx",
    "jsx"
  ],
  "output": "ts" as "ts" | "js" | "json",
  "outputDir": "src/i18n/locale",
  "withFileComment": true,
  "i18nFns": [
    "t",
    "$t",
    "i18n.t"
  ],
  "supportChineseKey": true,
  "languages": [
    "zh-CN",
    "en-US"
  ]
};
