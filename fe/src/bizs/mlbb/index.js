import { isEmptyValue } from "@/utils";
import { fetchPost } from "@/utils/http/onFetchRequest";

/**
 * @example 
 * openWebWhatsApp("我要参加组队预约新游活动！");
*/
export function openWebWhatsApp(messageText = "", phoneNumber = "6285873165264") {
  const encodedMessage = encodeURIComponent(messageText);
  const linkUrl = `https://wa.me/${phoneNumber}?text=${encodedMessage}`;
  // alert(`预备打开whatsapp，linkUrl为：${linkUrl}`);
  globalThis.location.href = linkUrl;
}

export function openPointLink(linkUrl = "") {
  globalThis.location.href = linkUrl;
}

// 调转到商店连接
export function openLinkStore(androidPackageName, iosAppId) {
  if (isAndroid()) {
    globalThis.location.href = `https://play.google.com/store/apps/details?id=${androidPackageName}`;
  } else if (isIOS()) {
    globalThis.location.href = `https://apps.apple.com/app/${iosAppId}`;
  } else {
    throw 'version error';
  }
}
/**
 * 尝试从Web打开MLBB游戏，如果没有安装则跳转到应用商店下载。
 * 
 * @param {string} path - 应用内的路径，例如'settings'
 * @param {string} androidPackageName - Android应用的包名
 * @param {string} iosAppId - iOS 应用的App Store ID
 * @param {string} schemeName - 自定义URL scheme，MLBB为'mlbb'
 */
export function openPlatformApp(path = '', androidPackageName = 'com.mobile.legends', iosAppId = 'id1160056295', schemeName = 'mlbb') {
  if (isAndroid()) {
    // Android下构建Intent URL
    const appLink = `intent://${path}#Intent;scheme=${schemeName};package=${androidPackageName};S.browser_fallback_url=${encodeURIComponent(`https://play.google.com/store/apps/details?id=${androidPackageName}`)};end`;
    tryOpenApp(appLink, `https://play.google.com/store/apps/details?id=${androidPackageName}`);
  } else if (isIOS()) {
    // iOS下使用自定义URL scheme打开应用，无法打开则跳转App Store
    const appLink = `${schemeName}://${path}`;
    tryOpenApp(appLink, `https://apps.apple.com/app/${iosAppId}`);
  } else {
    alert('This function is only supported on mobile devices.');
    console.log('This function is only supported on mobile devices.');
  }
}

/**
 * 数据埋点
 */
export function fetchTrackingPoint(configs = {}, apiUrl) {
  return new Promise((resolve, reject) => {
    const params = {
      // "fp": configs.fp,
      "type": "event",                // 必填  event | error   
      // "ffp": "",                   // 邀请当前用戶唯一标识
      // "fzoneid": "",               // 邀请当前用戶 zoneid  适用游戏网⻚活动
      // "froleid": "",               // 邀请当前用戶 roleid  适用游戏网⻚活动
      "data": configs.data
    };

    Object.entries(configs).forEach(([key, value], idx) => {
      params[key] = value;
    });

    fetchPost(apiUrl, {
      data: {
        "v2": JSON.stringify(params)
      }
    }, {
      headers: {
        "Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
        "urlencoded": 1
      }
    }).then(resp => {
      console.log(resp);
      resolve(resp);
    }).catch(err => {
      console.log(err);
      reject(err);
    });
  })
  // if (configs.ip) {
  //   params.data.ip = configs.ip;
  // }
  // if (configs.page) {
  //   params.data.page = configs.page;
  // }
  // if (userInfoData.zoneid) {
  //   params.zoneid = userInfoData.zoneid;  // 传递ML账号的roleid(如当下可获得roleid，获取不到可不传)
  // }
  // if (userInfoData.roleid) {
  //   params.roleid = userInfoData.roleid;  // 传递ML账号的zoneid(如当下可获得zoneid，获取不到可不传)
  // }
}
/**
 * SDK 数据埋点
*/
export function fetchSDKTrackingPoint(configs = {}) {
  if (isEmptyValue(globalThis.trackOptions)) {
    globalThis.trackOptions = globalThis.trackOptions || [];
  }
  // const params = {};
  Object.entries(configs).forEach(([key, value], idx) => {
    // params[key] = value;
    globalThis.trackOptions.push([key, value]);
  });
  if (typeof MtTrack !== 'undefined' && MtTrack.track) {
    // 调用 MtTrack.track 上报事件
    MtTrack?.track(eventName, configs).then(() => {
      console.log(`事件 "${eventName}" 上报成功`, eventData);
      // resolve();
    }).catch((err) => {
      console.error(`事件 "${eventName}" 上报失败`, err);
      // reject(err);
    });
  } else {
    console.error('MtTrack SDK 未正确加载');
    // reject(new Error('MtTrack SDK 未正确加载'));
  }

  // MtTrack && MtTrack.track();
  // return globalThis.trackOptions.push(arguments)
}
/**
 * 尝试打开应用，如果失败则跳转到下载链接。
 * 
 * @param {string} appLink - 应用的深度链接或自定义URL
 * @param {string} fallbackLink - 回退的下载链接
 * @param {number} delay - 等待跳转的延迟时间（默认2秒）
 */
function tryOpenApp(appLink, fallbackLink, delay = 2000) {
  window.location.href = appLink;
  alert(`tryOpenApp-${appLink}`);
  alert(`fallbackLink-${fallbackLink}`);

  let timer = setTimeout(() => {
    // 如果页面依然可见，说明应用没有成功打开，跳转到回退链接
    if (document.visibilityState === 'visible') {
      window.location.href = fallbackLink;
    }
  }, delay);

  // 监听页面的可见性，如果用户在应用打开时切换回页面，取消跳转
  function cancelFallback() {
    if (document.visibilityState === 'visible') {
      clearTimeout(timer);
      document.removeEventListener('visibilitychange', cancelFallback);
    }
  }
  document.addEventListener('visibilitychange', cancelFallback);
}

/**
 * 检查是否为Android设备
 * @returns {boolean}
 */
function isAndroid() {
  return /Android/i.test(navigator.userAgent);
}

/**
 * 检查是否为iOS设备
 * @returns {boolean}
 */
function isIOS() {
  return /iPhone|iPad|iPod/i.test(navigator.userAgent);
}
