export default {
  // 扫描入口
  entryDirs: [
    'src/views',
    'src/pages',
    'src/components'
  ],

  // 允许扫描的文件类型
  extensions: ['vue', 'tsx', 'jsx'],

  // 输出格式：js | ts | json
  output: 'ts',

  // 输出语言包目录
  outputDir: 'src/i18n/locale',

  // 是否在文件头部写注释说明扫描来源
  withFileComment: true,

  // 匹配的国际化函数项
  i18nFns: ['t', '$t']
}