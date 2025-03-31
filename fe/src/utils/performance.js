/**
 * 存储setTimeout和setInterval的句柄
 * @type {Object.<number, Object>}
 */
let timeoutHandles = {};
let intervalHandles = {};
let handleCounter = 0;

/**
 * 使用requestAnimationFrame实现的setTimeout
 * @param {Function} callback - 要执行的回调函数
 * @param {number} delay - 延迟时间（毫秒）
 * @returns {number} 返回一个句柄ID，可用于清除定时器
 * @example
 * // 设置一个2秒后执行的定时器
 * const timeoutId = rafSetTimeout(() => {
 *   console.log('2秒后执行');
 * }, 2000);
 * 
 * // 如果需要取消
 * rafClearTimeout(timeoutId);
 */
export function rafSetTimeout(callback, delay) {
  const start = performance.now();
  const handleId = ++handleCounter;

  function tick(now) {
    if (!timeoutHandles[handleId]) return; // 已被清除

    const elapsed = now - start;
    if (elapsed >= delay) {
      callback();
      delete timeoutHandles[handleId];
    } else {
      timeoutHandles[handleId].rafId = requestAnimationFrame(tick);
    }
  }

  timeoutHandles[handleId] = { callback, delay, start };
  timeoutHandles[handleId].rafId = requestAnimationFrame(tick);
  return handleId;
}

/**
 * 清除由rafSetTimeout创建的定时器
 * @param {number} handleId - 由rafSetTimeout返回的句柄ID
 * @example
 * const timeoutId = rafSetTimeout(() => {
 *   console.log('这段代码不会执行');
 * }, 1000);
 * 
 * rafClearTimeout(timeoutId);
 */
export function rafClearTimeout(handleId) {
  if (timeoutHandles[handleId]) {
    cancelAnimationFrame(timeoutHandles[handleId].rafId);
    delete timeoutHandles[handleId];
  }
}

/**
 * 使用requestAnimationFrame实现的setInterval
 * @param {Function} callback - 要定期执行的回调函数
 * @param {number} delay - 时间间隔（毫秒）
 * @returns {number} 返回一个句柄ID，可用于清除定时器
 * @example
 * // 每秒执行一次
 * const intervalId = rafSetInterval(() => {
 *   console.log('每秒执行一次');
 * }, 1000);
 * 
 * // 5秒后停止
 * rafSetTimeout(() => {
 *   rafClearInterval(intervalId);
 * }, 5000);
 */
export function rafSetInterval(callback, delay) {
  const start = performance.now();
  const handleId = ++handleCounter;

  function tick(now) {
    if (!intervalHandles[handleId]) return; // 已被清除

    const elapsed = now - intervalHandles[handleId].lastTime;
    if (elapsed >= delay) {
      intervalHandles[handleId].lastTime = now;
      callback();
    }
    intervalHandles[handleId].rafId = requestAnimationFrame(tick);
  }

  intervalHandles[handleId] = {
    callback,
    delay,
    lastTime: start
  };

  intervalHandles[handleId].rafId = requestAnimationFrame(tick);
  return handleId;
}

/**
 * 清除由rafSetInterval创建的定时器
 * @param {number} handleId - 由rafSetInterval返回的句柄ID
 * @example
 * const intervalId = rafSetInterval(() => {
 *   console.log('定期执行');
 * }, 1000);
 * 
 * // 停止定期执行
 * rafClearInterval(intervalId);
 */
export function rafClearInterval(handleId) {
  if (intervalHandles[handleId]) {
    cancelAnimationFrame(intervalHandles[handleId].rafId);
    delete intervalHandles[handleId];
  }
}