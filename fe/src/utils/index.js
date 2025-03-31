/**
 * 格式化日期时间
 * @param {Date|string|number} [input=new Date()] - 输入的日期
 * @param {number} [offsetDays=0] - 日期偏移天数
 * @param {string} [format="YYYY-MM-DD hh:mm:ss"] - 输出格式
 * @returns {Object} 返回包含各种格式的日期时间对象
 * @example
 * // 格式化当前时间
 * const now = formatDateTime();
 * console.log(now.dayFormat); // "2024-03-21 15:30:45"
 * 
 * // 格式化指定日期并偏移天数
 * const future = formatDateTime("2024-03-21", 7, "YYYY-MM-DD");
 * console.log(future.dayFormat); // "2024-03-28"
 */
export function formatDateTime(input = new Date, offsetDays = 0, format = "YYYY-MM-DD hh:mm:ss",) {
  // Helper to pad zeroes for date formatting
  const pad = (num, size) => num.toString().padStart(size, '0');

  // Convert input to a Date object
  let date;
  if (input instanceof Date) {
    date = new Date(input.getTime());
  } else if (typeof input === 'number') {
    date = new Date(input);
  } else {
    // Parse the input and check if time is missing
    date = new Date(input);
    if (input.trim().length <= 10) {
      // Time is missing, add current time
      const now = new Date();
      date.setHours(now.getHours(), now.getMinutes(), now.getSeconds(), now.getMilliseconds());
    }
  }

  // Apply the day offset
  date.setDate(date.getDate() + offsetDays);

  // Extract components
  const year = date.getFullYear();
  const month = date.getMonth() + 1;
  const day = date.getDate();
  const hours = date.getHours();
  const minutes = date.getMinutes();
  const seconds = date.getSeconds();
  const monthDayMaxCount = new Date(year, month, 0).getDate(); // Days in month

  // Format output based on the provided format
  const replacements = {
    'YYYY-MM-DD': `${year}-${pad(month, 2)}-${pad(day, 2)}`,
    '年-月-日': `${year}年${pad(month, 2)}月${pad(day, 2)}日`,
    'MM-DD': `${pad(month, 2)}-${pad(day, 2)}`,
    'YYYY-MM-DD hh:mm:ss': `${year}-${pad(month, 2)}-${pad(day, 2)} ${pad(hours, 2)}:${pad(minutes, 2)}:${pad(seconds, 2)}`,
    'hh:mm:ss': `${pad(hours, 2)}:${pad(minutes, 2)}:${pad(seconds, 2)}`,
    'mm:ss': `${pad(minutes, 2)}:${pad(seconds, 2)}`
  };

  const formattedDay = replacements[format] || input;

  // Today's date
  const today = new Date();
  const todayFormatted = `${today.getFullYear()}-${pad(today.getMonth() + 1, 2)}-${pad(today.getDate(), 2)} ${pad(today.getHours(), 2)}:${pad(today.getMinutes(), 2)}:${pad(today.getSeconds(), 2)}`;
  const todayTimestamp = today.getTime();

  // Create the response object
  const resultOptions = {
    yyyy: year.toString(),
    mm: pad(month, 2),
    dd: pad(day, 2),
    hh: pad(hours, 2),
    _mm: pad(minutes, 2),
    ss: pad(seconds, 2),
    timestamp: date.getTime(),
    dayDateInit: `${year}-${pad(month, 2)}-${pad(day, 2)} 00:00:00`,
    dayDate: `${year}-${pad(month, 2)}-${pad(day, 2)} ${pad(hours, 2)}:${pad(minutes, 2)}:${pad(seconds, 2)}`,
    dayFormat: formattedDay,
    _date: formattedDay.substring(0, 10),
    _time: formattedDay.substring(11),
    monthMax: monthDayMaxCount,
    date: new Date(date),
    dateInit: new Date(year, month - 1, day, 0, 0, 0),
    today: todayFormatted,
    todayTimestamp: todayTimestamp
    // dateFlatInit: ""
  };
  // resultOptions.dateFlatInit = resultOptions.dayDateInit.replace(/[\:\-\s]+/gi, "");

  return resultOptions;
}

/**
 * 监听元素尺寸变化并根据阈值返回对应值
 * @param {HTMLElement} element - 要监听的元素
 * @param {number[]} sizes - 尺寸阈值数组
 * @param {any[]} values - 对应阈值的返回值数组
 * @param {Function} cb - 回调函数
 * @param {string} [dimension="width"] - 监听的维度
 * @example
 * const element = document.querySelector('.responsive-element');
 * adaptElementSize(
 *   element,
 *   [1200, 992, 768],
 *   ['large', 'medium', 'small'],
 *   (size) => console.log(`Current size: ${size}`),
 *   'width'
 * );
 */
export function adaptElementSize(element, sizes = [], values = [], cb = () => { }, dimension = "width") {
  // 确保sizes和values数组按倒叙排列
  sizes = sizes.slice().sort((a, b) => b - a);
  values = values.slice().sort((a, b) => b - a);

  if (!element, sizes.length !== values.length) {
    throw "函数的参数设置错误";
  }

  // 创建一个promise，以便我们可以返回观察结果
  // return new Promise((resolve, reject) => {});
  // 创建 ResizeObserver 实例并提供处理变动的回调函数
  const observer = new ResizeObserver(entries => {
    for (let entry of entries) {
      const { width, height } = entry.contentRect;
      const currentSize = dimension === 'width' ? width : height;

      // 如果当前尺寸小于sizes数组中的最小值，则使用values数组中的最小值
      if (currentSize < sizes?.at(-1)) {
        cb(values?.at(-1));
        return;
      }

      // 查找符合当前尺寸的阈值，并获取对应的值
      for (let i = 0; i < sizes.length; i++) {
        if (currentSize >= sizes[i]) {
          // console.log(`当前尺寸是：${currentSize}，对应的值是：${values[i]}`);
          cb(values[i]);
          console.log(values[i]);
          return; // 如果找到匹配的尺寸，立即返回
        }
      }

      // 如果所有阈值都不满足，可能需要有个默认行为
      // console.log(`当前尺寸是：${currentSize}，没有对应的值`);
      cb(null); // 或者可以选择 reject(new Error("No matching size found"));
    }
  });

  // 开始观察目标节点
  observer.observe(element);

  // 提供一个方法来停止观察
  // 注意：如果需要在外部停止观察，你需要将observer实例存储在更高的作用域中
  // return () => observer.disconnect();

}

/**
 * 监听窗口尺寸变化
 * @param {number[]} sizes - 尺寸阈值数组
 * @param {any[]} values - 对应阈值的返回值数组
 * @param {Function} cb - 回调函数
 * @param {string} [dimension="width"] - 监听的维度
 * @example
 * adaptWindow(
 *   [1920, 1200, 768],
 *   [3, 2, 1],
 *   (value) => console.log(`Layout columns: ${value}`),
 *   'width'
 * );
 */
export function adaptWindow(sizes, values, cb = () => { }, dimension = "width") {
  // 注册尺寸变化的事件处理器
  window.removeEventListener('resize', handleResize, false);
  window.addEventListener('resize', handleResize, false);

  // 处理窗口尺寸变化的函数
  function handleResize() {
    // 获取当前的宽度或高度
    const currentSize = dimension === 'width' ? window.innerWidth : window.innerHeight;

    // 查找符合当前尺寸的阈值，并获取对应的值
    for (let i = 0; i < sizes.length; i++) {
      if (currentSize >= sizes[i]) {
        console.log(`当前尺寸是：${currentSize}，对应的值是：${values[i]}`);
        // 这里你可以替换为你需要执行的逻辑
        return values[i];
      }
    }

    // 如果所有阈值都不满足，可能需要有个默认行为
    console.log(`当前尺寸是：${currentSize}，没有对应的值`);
    // 这里你可以替换为你需要执行的逻辑
    return null;
  }

  // 立即执行一次，以处理页面加载时的尺寸
  let result = handleResize();
  cb(result, handleResize);
  // window.removeEventListener('resize', handleResize, false);
}

/**
 * 检查值是否为`空`，包含：undefined、null、空字符串、{}、[];
 *
 * @param {undefined|null|string|object|array} value 传入的值。
 * @returns {boolean} 返回一个布尔值，为 `true` 是拦截成功。
 */
export function isEmptyValue(value) {
  if ((value === undefined || value === null || value === '')) {
    return true;
  }
  if (typeof value === 'object' && !Object.keys(value).length) {
    return true;
  }
  if ((Array.isArray(value) && !value.length)) {
    return true;
  }
  return false;
}

export function sleep(time = 1000) {
  return new Promise(resolve => {
    requestTimeout(resolve, time);
  });
}

export function safeStringify(opts = {}) {
  if (typeof opts !== 'object') {
    return "";
  }
  return JSON.stringify(opts);
}

// 安全的JSONparse，避免报错导致白屏。
export function safeJSONparse(value, defaultValue = {}) {
  try {
    if (typeof value === "object") {
      return value;
    }
    if (typeof value === "string") {
      return JSON.parse(value);
    }
  } catch (e) {
    console.error('JSON解析错误:', e);
    // 返回默认值或执行其他错误处理
    return defaultValue;
  }
}

// 获得到链接上？号后面的参数
export function queryParams(key, url) {
  const querys = url ?? window.location.search;
  const queryParams = {};
  // 去掉开头的 '?'，如果存在的话
  const queryString = querys.startsWith('?') ? querys.slice(1) : querys;
  // 按 '&' 分割成每个键值对
  const pairs = queryString.split('&');

  for (const pair of pairs) {
    // 跳过空字符串，避免出错
    // if (!pair) continue;
    // // 用 '=' 分割键和值
    const [fieldKey, fieldValue] = pair.split('=');
    // // 处理解码，保留 '+'
    // const decodedKey = decodeURIComponent(key);
    // const decodedValue = decodeURIComponent((value || '').replace(/\+/g, ' '));

    // 存入对象
    queryParams[fieldKey] = decodeURIComponent(fieldValue);

    // 处理一些特殊值
    if (queryParams[fieldKey] === "null") {
      queryParams[fieldKey] = null;
    }
    if (queryParams[fieldKey] === "undefined") {
      queryParams[fieldKey] = undefined;
    }
    if (queryParams[fieldKey] === "true") {
      queryParams[fieldKey] = true;
    }
    if (queryParams[fieldKey] === "false") {
      queryParams[fieldKey] = false;
    }
  }

  if (!key) {
    return queryParams;
  }

  return queryParams[key];
}

/**
 * 按顺序获取URL参数，保持参数的原始顺序
 * @param {string} [url] - URL字符串，如果不提供则使用当前页面URL
 * @returns {Array} 返回参数数组，每个元素包含key和value
 * @example
 * // URL: "http://example.com?name=John&age=25&title=Hello World"
 * const params = queryParamsInOrder();
 * console.log(params);
 * // 输出: [
 * //   { key: "name", value: "John" },
 * //   { key: "age", value: "25" },
 * //   { key: "title", value: "Hello World" }
 * // ]
 */
export function queryParamsInOrder(url) {
  // 获取查询字符串
  const querys = url ?? window.location.search;

  // 如果没有查询参数，返回空数组
  if (!querys) return [];

  // 去掉开头的 '?'，如果存在的话
  const queryString = querys.startsWith('?') ? querys.slice(1) : querys;

  // 如果查询字符串为空，返回空数组
  if (!queryString) return [];

  // 按 '&' 分割成参数数组
  return queryString.split('&').map(pair => {
    // 查找第一个 '=' 的位置
    const firstEqualIndex = pair.indexOf('=');

    // 如果没有 '='，返回整个字符串作为key，value为空
    if (firstEqualIndex === -1) {
      return {
        [pair]: value,
        // key: pair,
        // value: ''
      };
    }

    // 使用第一个 '=' 来分割key和value
    const key = pair.slice(0, firstEqualIndex);
    const value = pair.slice(firstEqualIndex + 1);

    // 返回解码后的key和value
    return {
      [key]: decodeURIComponent(value),
      // value: decodeURIComponent(value)
    };
  });
}

// 生成唯一id
export function uuid() {
  return Math.random().toString(16).substring(2);
}
/**
 * 每一个 x，它生成一个 0 到 15 之间的随机数。
 * 每一个 y，它生成一个 8、9、A 或 B 中的一个值（根据 UUID v4 的规范）。
 */
export function uuidv4() {
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function (c) {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

/**
 * 复制文本
 * @param {string} text - 要复制的文本
 * @param {Function} successCallback - 复制成功回调
 * @param {Function} failCallback - 复制失败回调
 */
export function copyText(text = "", successCallback = () => { }, failCallback = () => { }) {
  // 使用旧方法向下兼容
  let textarea = document.createElement('textarea');
  textarea.value = text;
  textarea.style.position = 'fixed'; // 防止页面滚动
  textarea.style.left = '-9999999px';
  textarea.style.top = "-9999999px";
  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();
  try {
    let successful = document?.execCommand('copy');
    if (successful) {
      // alert('文本已复制到剪贴板');
      successCallback();
    } else {
      // alert('复制失败，请手动复制');
      failCallback({ message: '复制失败，请手动复制' });
    }
  } catch (err) {
    // alert('当前浏览器不支持自动复制，请手动复制');
    failCallback(err);
  }
  document.body.removeChild(textarea);
  // if (navigator?.clipboard && navigator?.clipboard?.writeText) {
  //   // 使用现代 API 进行复制
  //   navigator.clipboard.writeText(text).then(() => {
  //     // alert('文本已复制到剪贴板');
  //     successCallback();
  //   }).catch((err) => {
  //     // alert('复制失败，请手动复制');
  //     failCallback(err);
  //   });
  // }
}