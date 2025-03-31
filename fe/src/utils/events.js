/**
 * 存储所有事件回调函数的对象
 * @type {Object.<string, Array<Function>>}
 */
const eventStores = {};

/**
 * 添加事件监听器
 * @param {string} keyname - 事件名称
 * @param {Function} fn - 回调函数
 * @throws {string} 当缺少参数或参数类型错误时抛出错误
 * @example
 * // 添加一个窗口大小改变的事件监听
 * postEvent('windowResize', () => {
 *   console.log('Window size changed');
 * });
 * 
 * // 添加一个自定义事件监听
 * postEvent('customEvent', (data) => {
 *   console.log('Custom event triggered:', data);
 * });
 */
export function postEvent(keyname = "", fn) {
  if (!keyname || keyname && typeof fn !== "function") {
    throw "缺少实参或实参类型错误";
  }
  if (!eventStores[keyname]) {
    eventStores[keyname] = [];
  }

  eventStores[keyname].push(fn);
}

/**
 * 监听窗口大小改变和屏幕方向改变事件
 * @param {string} [key="windowResizeAndOrientationChange"] - 事件存储的键名
 * @example
 * // 添加窗口大小改变监听
 * postEvent('windowResizeAndOrientationChange', () => {
 *   console.log('Window resized or orientation changed');
 * });
 * onEventWinResize();
 * 
 * // 使用自定义键名
 * postEvent('myResizeEvent', () => {
 *   console.log('Custom resize handler');
 * });
 * onEventWinResize('myResizeEvent');
 */
export function onEventWinResize(key = "windowResizeAndOrientationChange") {
  if (!eventStores[key]) {
    eventStores[key] = [];
  }

  // 定义一个处理窗口变化的函数
  function handleWindowResize() {
    eventStores[key].map(cb => {
      if (typeof cb === "function") {
        cb();
      }
    });
  }

  if (!isOrientationChangeEventSupported("orientationchange")) {
    // 添加事件监听器，在窗口尺寸改变时重新计算
    globalThis.removeEventListener('resize', handleWindowResize, false);
    globalThis.addEventListener('resize', handleWindowResize, false);
  }
  // 适配移动设备的屏幕翻转
  globalThis.removeEventListener('orientationchange', handleWindowResize, false);
  globalThis.addEventListener('orientationchange', handleWindowResize, false);
}

/**
 * 检查是否支持特定的事件API
 * @param {string} [apiname=""] - 要检查的API名称
 * @returns {boolean} 如果支持该API返回true，否则返回false
 * @example
 * // 检查是否支持orientationchange事件
 * if (isOrientationChangeEventSupported('orientationchange')) {
 *   console.log('Orientation change is supported');
 * }
 * 
 * // 检查是否支持其他API
 * if (isOrientationChangeEventSupported('someOtherAPI')) {
 *   console.log('someOtherAPI is supported');
 * }
 */
export function isOrientationChangeEventSupported(apiname = "") {
  return apiname in globalThis;
}

/**
 * 监听设备方向改变事件（已注释掉的备选实现）
 * @deprecated 使用 onEventWinResize 替代
 * @example
 * // 不推荐使用此方法，请使用 onEventWinResize
 * // onWinOrientation();
 */
// export function onWinOrientation() {
//   if (!events["windowResizeAndOrientationChange"]) {
//     events["windowResizeAndOrientationChange"] = [];
//   }
//   ...
// }

