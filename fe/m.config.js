const path = require('path');
// const webpackModuleFederationUrl = {
//   "0": "http://localhost:8091",
//   "1": "/events/mcgg2025wa" // "//tianyuanminiparams.oss-cn-beijing.aliyuncs.com/projects"
// }[+global["$isProduction"]];

// console.log('webpackModuleFederationUrl', webpackModuleFederationUrl);


function getFormattedDateTime() {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, '0');  // 补齐2位
  const day = String(now.getDate()).padStart(2, '0');         // 补齐2位
  const hours = String(now.getHours()).padStart(2, '0');      // 补齐2位
  const minutes = "00"; // String(now.getMinutes()).padStart(2, '0');  // 补齐2位
  const seconds = "00"; //String(now.getSeconds()).padStart(2, '0');  // 补齐2位

  return `${year}${month}${day}${hours}${minutes}${seconds}`;
}

module.exports = {
  assetsPublicPath: "/events/mcgg2025wa/", //"//tianyuanminiparams.oss-cn-beijing.aliyuncs.com/projects/mlbbdeeplink/bundles/",
  assetsPath: `myBundleCode/${getFormattedDateTime()}`,
  base: {
    output: {
      uniqueName: "myApp"
    },
    alias: {
      '@': getAppnamePath('/src'),
      '@assets': getAppnamePath('/src/assets'),
      '@utils': getAppnamePath('/src/utils'),
      '@styles': getAppnamePath('/src/styles'),
      '@bizs': getAppnamePath('/src/bizs'),
      '@com': getAppnamePath('/src/components'),
      '@pages': getAppnamePath('/src/pages')
    },
    rules: [],
    plugins: [],
  },
  dev: {
    server: "http",
    open: false,
    hot: !true,
    port: "auto",
    proxyTable: {
      '/YUI/**': {
        target: 'http://localhost:8091',
        secure: false,
        changeOrigin: true
      },
      '/resources/**': {
        target: 'http://127.0.0.1:9001',
        secure: false,
        changeOrigin: true
      },
      '/r': {
        target: 'https://api.mobilelegends.com',
        secure: false,
        changeOrigin: true
      },
      '/events/mcgg2025wa/activity': {
        target: 'http://prod-mlbb-api.wysoftware.top',
        ws: true,
        secure: false,
        changeOrigin: true
      }
    },
    historyApiFallback: {
      rewrites: [{ from: /.*/g, to: '/index.html' }],
    },
    rules: [{
      test: /\.(jsx)$/,
      enforce: "pre", // 确保代码被其他loader处理之前执行。
      use: [
        {
          loader: "lint-reactcode-loader",
          options: {
            hookPath: "@hooks/usePrimitiveState",
            proxyMaxCount: 3,
            stateMaxCount: 5
          }
        }
      ],
      exclude: [
        getAppnamePath("./node_modules"),
        getAppnamePath("./src/common")
      ],
      include: [
        getAppnamePath("./src/components")
      ]
    }],
    plugins: []
  },
  pro: {
    rules: [],
    plugins: []
  }
};

function getAppnamePath(name) {
  return path.join(__dirname, name);
};
