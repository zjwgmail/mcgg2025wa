module.exports = {
  plugins: {
    'postcss-pxtorem': {
      rootValue: 20, // 根据你的基准大小调整
      propList: ['*', '!--*'], // 可以转换的属性列表, '*' 代表所有属性都要转换；!--是排除 CSS 变量
      unitPrecision: 5, // 转换后的小数位数
      selectorBlackList: ["html", /^\.adm-/], // 不进行px转换的选择器
      replace: true, // 是否直接替换掉原来的值，而不是添加备用值
      mediaQuery: true, // 允许在媒体查询中转换px
      minPixelValue: 0 // 设置要替换的最小像素值
    }
  }
};
