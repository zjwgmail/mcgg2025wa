/**
 * mobile adaption
*/

/**
 * 根据屏幕宽度计算自适应值
 * @param {number} [px=1] - 基准像素值
 * @param {number} [sizeWidth=750] - 设计稿宽度
 * @param {boolean} [isMax=true] - 是否限制最大宽度
 * @returns {number} 返回计算后的值
 * @example
 * // 计算基于750px设计稿的自适应值
 * const adaptedValue = adaptionWebViewPort(100); // 如果当前屏幕是375px，返回50
 * 
 * // 不限制最大宽度
 * const unlimitedValue = adaptionWebViewPort(100, 750, false);
 */
export function adaptionWebViewPort(px = 1, sizeWidth = 750, isMax = true) {
  let screenWidth = globalThis.innerWidth; // 获取当前窗口宽度
  if (isMax) {
    screenWidth = screenWidth >= sizeWidth ? sizeWidth : screenWidth;
  }
  const value = screenWidth / sizeWidth * px; // 示例计算，根据需要修改
  // console.log('Calculated value based on width:', value);
  return value;
}

/**
 * 查询DOM元素
 * @param {string} name - CSS选择器
 * @returns {Element|null} 返回匹配的第一个元素
 * @example
 * const header = queryElement('.header');
 * const menu = queryElement('#main-menu');
 */
export function queryElement(name = "") {
  return document.querySelector(name);
}

/**
 * 设置元素的样式和属性
 * @param {string} name - CSS选择器
 * @param {Object} styles - 样式对象
 * @param {Object} attrs - 属性对象
 * @example
 * // 设置元素样式和属性
 * postElement('html', {
 *   'font-size': '16px',
 *   'background-color': '#fff'
 * }, {
 *   'data-theme': 'light'
 * });
 */
export function postElement(name = "", styles = {}, attrs = {}) {
  let ele = queryElement(name);
  let cssList = [];

  Object.entries(styles).forEach(([key, value], idx) => {
    cssList.push(`${key}: ${value}`);
  });
  ele.setAttribute("style", cssList.join(";"));

  Object.entries(attrs).forEach(([key, value], idx) => {
    ele.setAttribute(key, value);
  });
}

/**
 * 监听页面可见性变化
 * @param {Function} hiddenCallback - 页面隐藏时的回调
 * @param {Function} visibilityCallback - 页面可见时的回调
 * @example
 * handleVisibilityChange(
 *   () => console.log('页面隐藏了'),
 *   () => console.log('页面可见了')
 * );
 */
export function handleVisibilityChange(hiddenCallback = () => { }, visibilityCallback = () => { }) {
  const visibilitychange = () => {
    if (document.hidden) {
      hiddenCallback();       // 页面不可见时
    } else {
      visibilityCallback();  // 页面可见时
    }
  };

  document.removeEventListener('visibilitychange', visibilitychange, false);
  document.addEventListener('visibilitychange', visibilitychange, false);
}

/**
 * 使用requestAnimationFrame执行动画序列
 * @param {HTMLElement} element1 - 第一个动画元素
 * @param {HTMLElement} element2 - 第二个动画元素
 * @param {Array<Array>} animations - 动画配置数组
 * @param {boolean} [loop=true] - 是否循环播放
 * @example
 * const animations = [
 *   [{
 *     name: 'fadeIn',
 *     duration: 1,
 *     'timing-function': 'ease',
 *     count: 1,
 *     delay: 0
 *   }]
 * ];
 * 
 * executeAnimationsWithTwoElements(
 *   document.querySelector('.element1'),
 *   document.querySelector('.element2'),
 *   animations,
 *   true
 * );
 */
export function executeAnimationsWithTwoElements(element1, element2, animations, loop = true) {
  let animationDataIndex = 0; // 当前正在执行的动画数据项索引
  let elements = [element1, element2]; // 两个展示动画的元素
  /**
   * 执行指定元素上的动画组（两步动画：如[0] 和 [1]）
   * @param {HTMLElement} element - 需要应用动画的元素
   * @param {Array} animationGroup - 一组动画（如 [0] 和 [1]）
   * @param {Function} onComplete - 当前组动画执行完毕后的回调
   */
  function executeAnimationGroup(element, animationGroup, onComplete) {
    let animationIndex = 0;
    let startTime = null; // 用于记录每个动画的开始时间

    // 获取动画持续时间（以毫秒为单位）
    const getAnimationDuration = (animation) => animation.duration * 1000;

    // 获取动画延迟时间（以毫秒为单位）
    const getAnimationDelay = (animation) => animation.delay * 1000;

    // 使用 requestAnimationFrame 执行动画帧
    function executeNextAnimationFrame(timestamp) {
      if (!startTime) startTime = timestamp; // 第一次调用时初始化时间

      const currentAnimation = animationGroup[animationIndex];
      const elapsedTime = timestamp - startTime; // 计算当前经过的时间
      const totalDelay = getAnimationDelay(currentAnimation);
      const totalDuration = getAnimationDuration(currentAnimation);

      // 判断是否该开始动画
      if (elapsedTime >= totalDelay) {
        // 应用 CSS 动画样式
        element.style.animation = `${currentAnimation.name} ${currentAnimation.duration}s ${currentAnimation["timing-function"]} ${currentAnimation.count}`;

        // 如果动画执行完毕，清除样式并执行下一个动画
        if (elapsedTime >= totalDelay + totalDuration) {
          element.style.animation = ''; // 清除动画样式
          animationIndex++; // 移动到下一个动画

          if (animationIndex < animationGroup.length) {
            // 继续执行下一个动画
            startTime = null; // 重置开始时间
            requestAnimationFrame(executeNextAnimationFrame);
          } else {
            // 当前动画组执行完毕，调用 onComplete 回调
            onComplete();
          }
        } else {
          // 动画尚未完成，继续下一帧
          requestAnimationFrame(executeNextAnimationFrame);
        }
      } else {
        // 如果还没有达到延迟时间，继续等待
        requestAnimationFrame(executeNextAnimationFrame);
      }
    }

    // 开始执行第一个动画
    requestAnimationFrame(executeNextAnimationFrame);
  }

  /**
   * 递归执行动画数据的每一项，使用两个元素交替展示动画，循环动画数据
   */
  function executeNextAnimation() {
    const currentElement = elements[animationDataIndex % 2]; // 交替使用 element1 和 element2
    const currentAnimationGroup = animations[animationDataIndex % animations.length]; // 获取当前动画组

    // 执行当前元素的动画组
    executeAnimationGroup(currentElement, currentAnimationGroup, () => {
      // 动画组执行完毕，切换到下一个动画数据项
      animationDataIndex++;

      if (animationDataIndex < animations.length || loop) {
        if (animationDataIndex >= animations.length && loop) {
          animationDataIndex = 0; // 如果循环，重置动画数据索引
        }
        executeNextAnimation(); // 执行下一个动画数据项
      }
    });
  }

  // 启动动画执行
  executeNextAnimation();
}

/**
 * 异步加载外部资源
 * @param {('script'|'link')} tagName - 要创建的标签类型
 * @param {string} url - 资源URL
 * @returns {Promise<string>} 加载成功时返回成功消息
 * @example
 * // 加载CSS文件
 * loadResource('link', 'styles/theme.css')
 *   .then(msg => console.log(msg))
 *   .catch(err => console.error(err));
 * 
 * // 加载JavaScript文件
 * loadResource('script', 'js/plugin.js')
 *   .then(msg => console.log(msg))
 *   .catch(err => console.error(err));
 */
export function loadResource(tagName, url) {
  return new Promise((resolve, reject) => {
    // 创建标签
    let element;
    if (tagName === 'script') {
      element = document.createElement('script');
      element.src = url;
      element.async = true;
    } else if (tagName === 'link') {
      element = document.createElement('link');
      element.href = url;
      element.rel = 'stylesheet';
    } else {
      reject(new Error('Invalid tag name. Use "script" or "link".'));
      return;
    }

    // 设置加载成功和失败的回调
    element.onload = () => resolve(`Resource loaded: ${url}`);
    element.onerror = () => reject(new Error(`Failed to load resource: ${url}`));

    // 将元素添加到文档的 <head> 中
    document.head.appendChild(element);
  });
}





