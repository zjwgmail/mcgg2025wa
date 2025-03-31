# MCGG 2025 Web Application 前端代码逻辑文档

## 概述

本文档详细描述了 MCGG 2025 Web Application 的前端核心代码逻辑，重点分析了 `bootstrap.jsx` 文件的结构、功能模块及实现原理。该文件是应用程序的主入口，负责初始化、国际化处理、页面渲染以及各种用户交互功能。

## 代码结构

`bootstrap.jsx` 的代码结构清晰明了，主要包含以下几个部分：

1. **样式导入**: 按照规范有序导入全局样式文件
2. **依赖引入**: 导入 React 及第三方库依赖
3. **工具函数导入**: 导入自定义工具函数
4. **常量定义**: 定义全局常量
5. **主应用组件**: 定义主应用组件 `AppComponent`
6. **非落地页处理函数**: 定义 `onMounted` 函数处理非落地页场景
7. **应用渲染函数**: 定义 `generateElement` 函数用于渲染应用
8. **主入口函数**: 定义 `Main` 函数作为应用入口

## 核心功能模块

### 1. 国际化处理

应用实现了完整的国际化支持，包括：

```jsx
// 处理国际化数据
await handlerInternationalizationTransform(i18n.data);
await handlerInternationalizationTransform(i18n_mode1_5.data);

// 获取语言配置
viewData.l = queryInternationLang(viewData.queryParams.lang || "02", viewData.queryParams.mode);
viewData.ln = viewData.queryParams.lang || "02"; // 语言类型
viewData.arl = queryActivityRules(viewData.queryParams.lang || "02");
viewData.cct = queryCopyCDKTextLang(viewData.queryParams.lang || "02");
```

语言切换功能:
```jsx
function onSwitchLang(item, index) {
  viewData.l = queryInternationLang(item.value, viewData.queryParams.mode);
  viewData.ln = item.value;
  viewData.arl = queryActivityRules(item.value || "02");
  viewData.lIdx = index;
}
```

默认支持两种语言:
```jsx
const MAP_LANG = useRef([{
  "label": "EN",
  "value": "02"
}, {
  "label": "ID",
  "value": "04"
}]);
```

### 2. 响应式布局

应用使用了自定义的响应式布局方案，通过动态计算 HTML 根元素的 font-size 实现:

```jsx
const onViewportChange = () => {
  let adaptHtmlSize = adaptionWebViewPort(20, 1080, false);
  postElement('html', {
    'font-size': adaptHtmlSize + 'px'
  });
  return adaptHtmlSize;
}

// 监听窗口大小和方向变化
postEvent('windowResizeAndOrientationChange', () => {
  return onViewportChange();
});
onEventWinResize('windowResizeAndOrientationChange');
```

### 3. 页面路由与渲染

应用根据 URL 参数判断不同页面场景并进行相应渲染:

```jsx
// 根据URL参数判断是否为落地页
const landPage = queryParams("lp");      // 是否是落地页
const _getQueryParams = queryParams() ?? {};
if (landPage == 1) {
  generateElement();  // 落地页场景
} else {
  onMounted(_getQueryParams);       // 非落地页场景
}
```

### 4. 数据追踪与埋点

应用实现了完整的用户行为追踪和数据埋点功能:

```jsx
// 埋点上报
onCareOfIt(`reward${viewData.queryParams?.mode}PageExposure`);

// 埋点函数
async function onCareOfIt(behavior = "") {
  if (!viewData.queryParams?.mode) { 
    return;
  }
  fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
    "proj": "mcgg",
    "act_type": "mcgg2025wa",
    "behavior": behavior,
    "lang": viewData.queryParams?.lang || "02",  //语言 01中文 02英语 03马来语
    "channel": viewData.queryParams?.channel || "", //渠道
    "url": globalThis.location.href
  });
}
```

### 5. 用户交互功能

#### 5.1 复制游戏码

```jsx
// 复制游戏码
function onHandlerGameCode() {
  onCareOfIt(`reward${viewData.queryParams?.mode}ButtonClick`);
  copyText(viewData.queryParams?.cdk ?? "", () => {
    if (viewData.queryParams.mode == 3 || viewData.queryParams.mode == 8) {
      viewData.copyPopupVisible = true;
    }
  });
}
```

#### 5.2 弹窗控制

```jsx
// 奖励弹窗关闭按钮
function codeClosePopup() {
  viewData.copyPopupVisible = false;
}
```

## 页面渲染逻辑

应用根据不同参数渲染不同的页面内容:

### 1. 活动规则页面

当 `viewData.queryParams.gpt == 11` 时渲染活动规则页面:

```jsx
// 界面呈现 - 邀请活动规则
if (viewData.queryParams.gpt == 11) {
  return (
    <section className={`f16 oh ${styles[`langStyle${viewData.ln}`]} ${styles.AppContainerWrapper} ${styles.InvitationActivityRulesWrapper}`}>
      {/* 页面内容... */}
    </section>
  )
}
```

### 2. 普通奖励页面

默认渲染普通奖励页面:

```jsx
return (
  <section className={`f16 oh ${styles[`langStyle${viewData.ln}`]} ${styles.AppContainerWrapper}`}>
    {/* 页面内容... */}
  </section>
);
```

## 非落地页场景逻辑

非落地页场景根据不同的参数执行不同的跳转逻辑:

### 1. 预约新游场景

```jsx
// 从 前往预约新游 进入
if (!!queryCode && gamePageType == 1) {
  // 发送助力接口
  fetchPost('/events/mcgg2025wa/activity/help', {
    param: queryCode
  }, {
    notTipBizCodeMsg: true
  }).then(resp => {
    // 上报数据并跳转
    fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
      // 追踪参数...
    });
    openPointLink('https://8ufa.adj.st/?adj_t=1kyuom1r_1k7aid97&adj_redirect_ios=https%3A%2F%2Fapps.apple.com%2Fus%2Fapp%2Fmagic-chess-go-go%2Fid6612014908%3Fppid%3D88f9f6ab-4be0-46c7-8ad0-f9bf2e82b312');
  });
}
```

### 2. 游戏内兑换场景

```jsx
// 从 前往游戏内兑换 进入规则页
if ([10, 11, '10', '11'].includes(gamePageType)) {
  let preUrl = `${globalThis.location.origin}${globalThis.location.pathname}`;
  globalThis.location.href = `${preUrl}?code=${queryCode}&gpt=${gamePageType}&lp=1`;
}
```

### 3. 开团场景

```jsx
// 开团人的条件，没有用户的gpt，但是存在code集结码，作为开团人状态。
if (!!queryCode && !gamePageType) {
  // 上报数据
  fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
    // 追踪参数...
  });

  // 跳转WhatsApp
  let msgLangConfig = queryWhatsppMessageLang(gameLang || "02");
  rafSetTimeout(() => {
    openWebWhatsApp(msgLangConfig?.message({ code: queryCode }));
  }, 300);
}
```

## 启动流程

应用启动流程:

1. 设置全局产品名称: `globalThis["$ProductionName"] = mcgg_202501171530`
2. 执行视口适配: `onViewportChange()`
3. 注册窗口大小变化监听
4. 判断是否为落地页并执行相应逻辑