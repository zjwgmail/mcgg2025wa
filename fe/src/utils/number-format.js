
/**
 * 将数字转换为千分位格式
 * @param {number|string} num - 需要格式化的数字或字符串
 * @param {number} [decimals] - 保留的小数位数，可选参数
 * @returns {string} 返回格式化后的字符串
 * @example
 * formatThousands(1234567.89, 2)  // 返回: "1,234,567.89"
 * formatThousands('1234567.89')    // 返回: "1,234,567.89"
 * formatThousands(1234567)         // 返回: "1,234,567"
 * formatThousands(1234.56789, 3)   // 返回: "1,234.568"
 * formatThousands(0.123, 2)        // 返回: "0.12"
 * formatThousands('')              // 返回: "0"
 * formatThousands(null)            // 返回: "0"
 * formatThousands('abc')           // 返回: "0"
 */
export function formatThousands(num, decimals) {
  // 处理空值或非数字情况
  if (!num && num !== 0) return '0';

  // 将输入转换为数字
  const number = typeof num === 'string' ? Number(num) : num;

  // 如果转换后不是有效数字，返回0
  if (isNaN(number)) return '0';

  // 如果指定了小数位数，进行四舍五入
  const formattedNum = typeof decimals === 'number'
    ? number.toFixed(decimals)
    : number.toString();

  // 分割整数部分和小数部分
  const [integerPart, decimalPart] = formattedNum.split('.');

  // 对整数部分添加千分位分隔符
  const formattedInteger = integerPart.replace(/\B(?=(\d{3})+(?!\d))/g, ',');

  // 组合整数部分和小数部分
  return decimalPart ? `${formattedInteger}.${decimalPart}` : formattedInteger;
}

/**
 * 将数字转换为带中文单位的字符串
 * @param {number|string} num - 需要转换的数字或字符串
 * @param {number} [decimals=2] - 保留的小数位数，默认2位
 * @param {boolean} [withUnit=true] - 是否显示单位，默认显示
 * @returns {string} 返回带单位的字符串
 * @example
 * formatChineseNumber(647424831.591157636)    // 返回: "6.47亿"
 * formatChineseNumber(1234567.89)             // 返回: "123.46万"
 * formatChineseNumber(1234.56)                // 返回: "1,234.56"
 * formatChineseNumber(647424831.59, 4)        // 返回: "6.4742亿"
 * formatChineseNumber(647424831.59, 2, false) // 返回: "6.47"
 * formatChineseNumber('')                     // 返回: "0"
 * formatChineseNumber('abc')                  // 返回: "0"
 */
export function formatChineseNumber(num, decimals = 2, withUnit = true) {
  // 处理空值或非数字情况
  if (!num && num !== 0) return '0';

  // 将输入转换为数字
  const number = typeof num === 'string' ? Number(num) : num;

  // 如果转换后不是有效数字，返回0
  if (isNaN(number)) return '0';

  // 定义单位和阈值
  const units = [
    { value: 1e8, unit: '亿' },
    { value: 1e4, unit: '万' },
    { value: 1, unit: '' }
  ];

  // 找到适合的单位
  const unitInfo = units.find(unit => Math.abs(number) >= unit.value);

  if (!unitInfo) return '0';

  // 计算转换后的数值
  const convertedNum = number / unitInfo.value;

  // 格式化数字
  const formattedNum = Number(convertedNum.toFixed(decimals));

  // 如果数字小于1万，使用千分位格式
  if (unitInfo.value === 1) {
    return formatThousands(formattedNum, decimals);
  }

  // 返回结果
  return withUnit
    ? `${formattedNum}${unitInfo.unit}`
    : formattedNum.toString();
}

/**
 * 在不同的内存单位之间进行转换
 * @param {number} value - 需要转换的值
 * @param {Array<string>} units - 单位数组，第一个是源单位，第二个是目标单位。支持的单位：'b', 'kb', 'mb', 'gb'
 * @returns {number} 转换后的值，保留2位小数
 * @example
 * convertMemoryUnit(1024, ['b', 'kb'])     // 返回: 1
 * convertMemoryUnit(1024, ['kb', 'mb'])    // 返回: 1
 * convertMemoryUnit(1024, ['mb', 'gb'])    // 返回: 1
 * convertMemoryUnit(1, ['gb', 'mb'])       // 返回: 1024
 * convertMemoryUnit(1024, ['kb', 'b'])     // 返回: 1048576
 * convertMemoryUnit(1.5, ['gb', 'mb'])     // 返回: 1536
 */
export function convertMemoryUnit(value, [fromUnit, toUnit], showUnit = true) {
  // 单位到字节的转换比率
  const UNIT_MAP = {
    'b': 1,
    'kb': 1024,
    'mb': 1024 * 1024,
    'gb': 1024 * 1024 * 1024
  };

  // 检查单位是否有效
  const validUnits = Object.keys(UNIT_MAP);
  fromUnit = fromUnit.toLowerCase();
  toUnit = toUnit.toLowerCase();

  if (!validUnits.includes(fromUnit) || !validUnits.includes(toUnit)) {
    throw new Error(`Invalid unit. Supported units are: ${validUnits.join(', ')}`);
  }

  // 先转换为字节
  const bytes = value * UNIT_MAP[fromUnit];

  // 再从字节转换为目标单位
  const resultValue = bytes / UNIT_MAP[toUnit];

  // 保留2位小数
  const result = Math.round(resultValue * 100) / 100;

  // 保留2位小数
  return showUnit ? `${result}${toUnit.toLocaleUpperCase()}` : result;
}