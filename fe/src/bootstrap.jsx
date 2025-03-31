/**
 * 导入全局样式文件
 * 按照规范要求的顺序导入样式文件
 */
import './styles/css-reset.less';
import './styles/css-modules.less';
import './styles/variables/variables.less';
import './styles/themes/main-theme.less';
import './styles/css-global.less';
import styles from './bootstrap.less';

import { useEffect, useRef } from 'react';
import { createRoot } from "react-dom/client";
import { useReactive } from 'ahooks';
// import { Toast } from 'antd-mobile';

// 导入工具函数
import { adaptionWebViewPort, postElement } from './utils/m';
import { onEventWinResize, postEvent } from './utils/events';
import { copyText, queryParams, queryParamsInOrder, safeJSONparse, safeStringify } from './utils';
import { handlerInternationalizationTransform, i18n, i18n_mode1_5, LANGUAGE_MODE, queryActivityRules, queryCopyCDKTextLang, queryInternationLang, queryWhatsppMessageLang } from './stores/i18n';
import { decrypt } from './utils/encry/rsa';
import { decomposeCode, fetchSDKTrackingPoint } from './bizs';
import { fetchPost } from './utils/http/onFetchRequest';
import { openLinkStore, openPointLink, openWebWhatsApp } from './bizs/mlbb';
import { rafSetTimeout } from './utils/performance';

const CONSTANT_OPTIONS = {
  "projectId": 2810196,
};

/**
 * MCGG 主应用组件
 * @param {Object} props - 组件属性
 */
function AppComponent(props = {}) {
  const MAP_LANG = useRef([{
    "label": "EN",
    "value": "02"
  }, {
    "label": "ID",
    "value": "04"
  }]);
  const viewData = useReactive({
    queryParams: {},
    l: {},
    ln: "02",
    lIdx: 0,
    copyPopupVisible: false
  });

  useEffect(() => {
    onLoad();

    return () => { }
  }, []);

  // 初始化函数
  async function onLoad() {
    // 处理国际化数据 - 3、8
    await handlerInternationalizationTransform(i18n.data);
    // 处理国际化数据 - 1、5
    await handlerInternationalizationTransform(i18n_mode1_5.data);

    viewData.queryParams = queryParams();
    // viewData.queryParams?.cdk = queryParamsInOrder()?.at(3);
    if (viewData.queryParams?.cdk) {
      viewData.queryParams.cdk = decrypt(viewData.queryParams.cdk); // 解密
    }

    viewData.l = queryInternationLang(viewData.queryParams.lang || "02", viewData.queryParams.mode); // 获得语言配置
    viewData.ln = viewData.queryParams.lang || "02"; // 语言类型
    viewData.arl = queryActivityRules(viewData.queryParams.lang || "02");
    viewData.cct = queryCopyCDKTextLang(viewData.queryParams.lang || "02");

    // 埋点上报
    onCareOfIt(`reward${viewData.queryParams?.mode}PageExposure`);
  }
  // 切换语言
  function onSwitchLang(item, index) {
    viewData.l = queryInternationLang(item.value, viewData.queryParams.mode);
    viewData.ln = item.value;
    viewData.arl = queryActivityRules(item.value || "02");
    viewData.lIdx = index;
  }
  // 复制游戏码
  function onHandlerGameCode() {
    onCareOfIt(`reward${viewData.queryParams?.mode}ButtonClick`);
    copyText(viewData.queryParams?.cdk ?? "", () => {

      if (viewData.queryParams.mode == 3 || viewData.queryParams.mode == 8) {
        // if (viewData.ln === "04") {}
        viewData.copyPopupVisible = true;
      }
    });
  }
  // 3人奖励、5人奖励和8人奖励地址,控制参数定义
  async function onCareOfIt(behavior = "") {
    if (!viewData.queryParams?.mode) { // rewardpage
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
  // 奖励弹窗关闭按钮
  function codeClosePopup() {
    viewData.copyPopupVisible = false;
  }

  // 界面呈现 - 邀请活动规则
  if (viewData.queryParams.gpt == 11) {
    return (
      <section className={`f16 oh ${styles[`langStyle${viewData.ln}`]} ${styles.AppContainerWrapper} ${styles.InvitationActivityRulesWrapper}`}>
        <head className="df headerBox">
          <div className="df logoBox">
            <div className={`oh mcggLogo`} title="MCGG Logo"></div>
            <div className={`oh mlbbLogo`} title="MLBB Logo"></div>
          </div>
          <div className="df tac langSwitchBox sphide">
            {
              MAP_LANG.current.map((item, index) => {
                return <span onClick={() => onSwitchLang(item, index)} key={index} className={`langBtn ${viewData.ln == item.value ? 'active' : ''}`}>{item.label}</span>
              })
            }
          </div>
        </head>
        <div className={`df ruleTitleBox ruleTitleBox${viewData.ln}`}></div>
        <div className={`oh m0a contentBox contentBox${viewData.queryParams?.mode}`}>
          <div className={`activeRuleContentBox`}>
            {
              viewData.arl?.activeRuleContent?.map((it, idx) => {
                return (
                  <>
                    <div key={idx} className="df itText">
                      <strong className="oh tac serial">{(idx + 1)}</strong>
                      <span className="val" dangerouslySetInnerHTML={{ __html: (typeof it.text === "function" ? it.text(viewData.queryParams) : it.text) }}></span>
                    </div>
                    {
                      it.imgs?.length ? it.imgs.map((imgItem, imgIdx) => {
                        return (
                          <img key={imgIdx} className="itImg" src={imgItem.url} />
                        )
                      }) : ""
                    }
                  </>
                )
              })
            }
          </div>
        </div>

        <div className={`df ruleTitleBox prizeTableTitleImg${viewData.ln}`}></div>
        <div className={`oh m0a contentBox prizeTableBox contentBox${viewData.queryParams?.mode}`}>
          <div className={`prizeTable`}>
            <table cellPadding={0} cellSpacing={0} border={0}>
              <thead>
                <tr>
                  {
                    viewData.arl?.activeRuleWinningInfoContent?.columns?.map((it, idx) => {
                      return <th className="th" key={idx}>{it.thTitle}</th>
                    })
                  }
                </tr>
              </thead>
              <tbody>
                {
                  viewData.arl?.activeRuleWinningInfoContent?.dataSource?.map((it, idx) => {
                    return <tr key={idx}>
                      <td>{it.whatsappAccount}</td>
                      <td>{it.whatsappName}</td>
                      <td>{it.winningPrize}</td>
                    </tr>
                  })
                }
              </tbody>
            </table>
          </div>
        </div>


      </section>
    )
  }

  return (
    <section className={`f16 oh ${styles[`langStyle${viewData.ln}`]} ${styles.AppContainerWrapper}`}>
      <head className="df headerBox">
        <div className="df logoBox">
          <div className={`oh mcggLogo`} title="MCGG Logo"></div>
          <div className={`oh mlbbLogo`} title="MLBB Logo"></div>
        </div>

        <div className="df tac langSwitchBox sphide">
          {
            MAP_LANG.current.map((item, index) => {
              return <span onClick={() => onSwitchLang(item, index)} key={index} className={`langBtn ${viewData.ln == item.value ? 'active' : ''}`}>{item.label}</span>
            })
          }
        </div>
      </head>
      <div className={`bannerBox bannerBox${viewData.queryParams?.mode}`}></div>
      <div className={`pr packageCodeBox packageCodeBox${viewData.queryParams?.mode} packageCodeBoxMode${viewData.queryParams?.mode}`}>
        <span className="pa df gameCode">{viewData.queryParams?.cdk}</span>
        {
          viewData.queryParams?.mode == 5 || viewData.queryParams?.mode == 1 ? (
            <a onClick={onHandlerGameCode} className="pa urlLink" href="https://r8qs.adj.st/appinvites/UI_CDKey?adj_t=1i7vuwle_1iz74412&adjust_deeplink=mobilelegends%3A%2F%2Fappinvites%2FUI_CDKey" rel="noopener noreferrer" target="_blank"></a>
          ) : (
            <span onClick={onHandlerGameCode} className="pa urlLink"></span>
          )
        }
      </div>
      <div className={`oh m0a contentBox contentBox${viewData.queryParams?.mode}`}>
        <head className="df titBox">
          <span className="iconL"></span>
          <span className="tac tit" dangerouslySetInnerHTML={{ __html: viewData.l?.activeRuleContentTitle }}></span>
          <span className="iconR"></span>
        </head>
        <div className={`activeRuleContentBox`}>
          {
            viewData.l?.activeRuleContent?.map((it, idx) => {
              return (
                <>
                  <div key={idx} className="df itText">
                    <strong className="oh tac serial">{(idx + 1)}</strong>
                    <span className="val" dangerouslySetInnerHTML={{ __html: (typeof it.text === "function" ? it.text(viewData.queryParams) : it.text) }}></span>
                  </div>
                  {
                    it.imgs?.length ? it.imgs.map((imgItem, imgIdx) => {
                      return (
                        <img key={imgIdx} className="itImg" src={imgItem.url} />
                      )
                    }) : ""
                  }
                </>
              )
            })
          }
        </div>
      </div>
      <div className={`pf df codeBoxPopupWrapper ${viewData.copyPopupVisible ? '' : 'sphide'}`}>
        <div className="oh pr codeBoxPopup">
          <span className="pa close" onClick={codeClosePopup}></span>
          <h1 className="tac popup-title">{viewData.cct?.title}</h1>
          <div className="oys tal popup-content">
            {
              viewData.cct?.content?.map((it, idx) => {
                return <p className={`cont-${idx}`} key={idx} dangerouslySetInnerHTML={{ __html: it.text }}></p>
              })
            }
          </div>
        </div>
      </div>
    </section>
  );
}
/**
 * 非落地页场景的初始化逻辑
 */
function onMounted(qp = {}) {
  // let queryParams = queryParams();
  let gamePageType = qp.gpt; // 游戏页面类型
  let queryCode = qp?.code;   // 开团集结码 (rallyCode)
  let gameChannel = decomposeCode(queryCode); // 投放渠道
  let gameLang = qp?.lang; // 语言

  // 从 前往预约新游 进入
  if (!!queryCode && gamePageType == 1) {
    alert(`从 前往预约新游 按钮进来的`);
    alert(`先执行上报逻辑，然后执行：跳转功能？跳转到任务链接地址（需要先进入到游戏内？一个页面地址？再返回时看到完成开团信息？）`); // 疑问TODO：链接到哪？
    // let decry = decrypt(decodeURIComponent(queryCode));
    // if (typeof decry === "string") {
    //   decry = safeJSONparse(decry) || {};
    // }
    // alert(`decry` + "," + decry + "," + decodeURIComponent(queryCode) + "," + decomposeCode(decry.rally_code));
    // alert(`集结码拆解：${safeStringify(decomposeCode(decry.rally_code))}`);
    // alert(`code解密：${safeStringify(decry)}`);
    alert(`code解码后：${decodeURIComponent(queryCode)}`);
    alert(`发送助力接口，入参-param->解码后：${queryCode}`);

    fetchPost('/events/mcgg2025wa/activity/help', {
      param: queryCode
    }, {
      notTipBizCodeMsg: true
    }).then(resp => {
      if (resp.code !== 200) {
        alert(`resp.code !== 200：${resp?.message}`);
        return;
      }
      alert(`模拟 前往预约新游 的操作已完成，准备跳转至whatsapp`);
      alert('/events/mcgg2025wa/activity/help--POST', resp);

      fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
        // "projectId": 2810196,
        "proj": "mcgg",
        "act_type": "mcgg2025wa",
        "behavior": "appointmentLinkExposure",
        "lang": qp?.lang || "02",  //语言 01中文 02英语 03马来语
        "channel": qp?.channel || "", //渠道
        "url": globalThis.location.href
      });

      openPointLink('https://8ufa.adj.st/?adj_t=1kyuom1r_1k7aid97&adj_redirect_ios=https%3A%2F%2Fapps.apple.com%2Fus%2Fapp%2Fmagic-chess-go-go%2Fid6612014908%3Fppid%3D88f9f6ab-4be0-46c7-8ad0-f9bf2e82b312');
      // openLinkStore('com.mobile.legends', 'id1160056295');
    }).catch(err => {
      alert('err', err?.message);
    });
    // boss.whatsapp.openWebWhatsApp(""); // boss.utils.rsa.decryptData(queryCode)
    return;
  }
  // 从 前往游戏内兑换 进入规则页？
  if ([10, 11, '10', '11'].includes(gamePageType)) {
    alert(`从 前往游戏内兑换 按钮进来的`);
    alert(`疑问TODO：1.需要执行上报逻辑？2.需要从web网页打开游戏？`); // 疑问TODO：1.需要执行上报逻辑？2.需要从web网页打开游戏？
    let preUrl = `${globalThis.location.origin}${globalThis.location.pathname}`;
    alert(`通过 领取方式：${preUrl}?code=${queryCode}&gpt=${gamePageType}&lp=1&cdk=??CDK1001 进入最终奖励领取页面。`);
    globalThis.location.href = `${preUrl}?code=${queryCode}&gpt=${gamePageType}&lp=1`;
    // alert(`模拟 进入游戏`);
    // boss.whatsapp.openPlatformApp('', 'com.mobile.legends', 'id1160056295');
    return;
  }
  // 开团人的条件，没有用户的gpt，但是存在code集结码，作为开团人状态。
  if (!!queryCode && !gamePageType) {
    alert(`进入渠道对应的投放页面。${JSON.stringify(qp)}，投放渠道：${gameChannel}。`);
    alert(`准备跳转whatsapp并且执行：我要参与MLBB组队活动，抽vivo手机、RM5000现金和限定皮肤等奖励！\n我的活动码：${queryCode}`);
    // 上报数据
    let decomposeCodeOptions = decomposeCode(qp?.code); // 开团集结码
    let lang = qp?.lang; // 语言
    fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
      "proj": "mcgg",
      "act_type": "mcgg2025wa",
      "behavior": "promotionLinkExposure",
      "lang": lang || "02",  //语言 01中文 02英语 03马来语
      "channel": decomposeCodeOptions?.channel || "", //渠道  a 端内-128通路 b 端内-邮件推送 c 端内-任务达人 d 端外-fb e 端外-ins f 端外-ua加热 g 端外-备用1 h 端外-备用2
      "url": window.location.href
    });

    let msgLangConfig = queryWhatsppMessageLang(gameLang || "02");
    alert(`对应发送whatsapp消息：${msgLangConfig?.message({ code: queryCode })}`);
    rafSetTimeout(() => {
      openWebWhatsApp(msgLangConfig?.message({ code: queryCode }));
    }, 300);
    return;
  }
}

/**
 * 生成并渲染React应用
 * 用于落地页场景
 */
function generateElement() {
  const container = createRoot(document.querySelector("#root"));
  container.render(<AppComponent />);
}
/**
 * 应用主入口函数
 * 负责初始化全局配置和响应式适配
 */
function Main() {
  // 设置全局产品名称
  globalThis["$ProductionName"] = `mcgg_202501171530`;

  /**
   * 视口变化处理函数
   * 根据屏幕宽度计算html的font-size，实现响应式布局
   * @returns {number} 计算后的font-size值
   */
  const onViewportChange = () => {
    let adaptHtmlSize = adaptionWebViewPort(20, 1080, false);
    postElement('html', {
      'font-size': adaptHtmlSize + 'px'
    });
    return adaptHtmlSize;
  }
  // 初始化执行一次
  onViewportChange();

  // 监听窗口大小和方向变化，动态调整布局
  postEvent('windowResizeAndOrientationChange', () => {
    return onViewportChange();
  });
  onEventWinResize('windowResizeAndOrientationChange');

  // 根据URL参数判断是否为落地页
  const landPage = queryParams("lp");      // 是否是落地页
  const _getQueryParams = queryParams() ?? {};
  if (landPage == 1) {
    generateElement();  // 落地页场景
  } else {
    onMounted(_getQueryParams);       // 非落地页场景
  }
}
// 启动应用
Main();

// 辅助函数
function alert() { }