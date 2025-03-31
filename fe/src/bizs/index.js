// 分解code，拿到拆解出来的对应字段
export function decomposeCode(code = "") {
  // 获得投放code的参数信息
  let InvCode = code; //boss.utils.queryParams("code") ?? "";
  let channel = InvCode.substring(0, 1);  // 渠道
  let langNum = InvCode.substring(1, 3);  // 语言
  let algebra = InvCode.substring(3, 5);  // 代数
  let playerCode = InvCode.substring(5);  // 玩家识别码 

  return {
    channel,
    langNum,
    algebra,
    playerCode
  }
}

// 打点
export function fetchSDKTrackingPoint(trackId = 2810196, configs = {}) {
  // setTrackOptions("projectId", 2810196);
  // setTrackOptions("debug", true);
  if (typeof MtTrack !== 'undefined' && MtTrack.track) {
    // 调用 MtTrack.track 上报事件
    MtTrack?.track(trackId, configs);
  } else {
    console.error('MtTrack SDK 未正确加载');
    // reject(new Error('MtTrack SDK 未正确加载'));
  }
}


