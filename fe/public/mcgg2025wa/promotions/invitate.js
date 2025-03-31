function openWebWhatsApp(messageText = "", phoneNumber = "6285873165264") {
  const encodedMessage = encodeURIComponent(messageText);
  const linkUrl = `https://wa.me/${phoneNumber}?text=${encodedMessage}`;
  // alert(`预备打开whatsapp，linkUrl为：${linkUrl}`);
  globalThis.location.href = linkUrl;
}

function openWebWhatsAppToFriendList(messageText = "") {
  const encodedMessage = encodeURIComponent(messageText);
  const linkUrl = `https://wa.me/?text=${encodedMessage}`;
  // alert(`预备打开whatsapp，linkUrl为：${linkUrl}`);
  globalThis.location.href = linkUrl;
}

function getQueryParams(key, url) {
  const querys = decodeURIComponent(url ?? window.location.search);
  const params = new URLSearchParams(querys);
  Object.keys(params).forEach((it) => {
    if (params[it] === "null") {
      params[it] = null;
    }
    if (params[it] === "undefined") {
      params[it] = undefined;
    }
    if (params[it] === "true") {
      params[it] = true;
    }
    if (params[it] === "false") {
      params[it] = false;
    }
  });
  if (!key) {
    const result = {};
    for (const [name, value] of params.entries()) {
      result[name] = value;
    }
    return result;
  }

  if (Array.isArray(key)) {
    const result = {};
    for (let i = 0, len = key.length; i < len; i++) {
      const it = key[i];
      result[it] = params.get(it);
    }
    return result;
  }
  return params.get(key);
}

async function fetchRequest(method, url, params = {}, configs = {}) {
  // 初始化请求配置
  const options = {
    method: method.toUpperCase(), // 将方法名转为大写
    headers: {
      'Content-Type': 'application/json', // 默认使用 JSON 形式
      ...configs.headers // 合并用户传入的头信息
    }
  };

  // 处理不同请求方法的参数
  if (method.toLowerCase() === 'get' || method.toLowerCase() === 'delete') {
    // 对于 GET 和 DELETE 请求，将参数拼接到 URL 上
    const urlParams = new URLSearchParams(params).toString();
    if (!!urlParams) {
      url = `${url}?${urlParams}`;
    }
  } else {
    // 对于 POST 和 PUT 请求，将参数作为请求体
    options.body = JSON.stringify(params);
  }

  try {
    const response = await fetch(url, options);
    // 处理 HTTP 状态码
    handleHttpStatus(response.status);
    // 检查响应状态
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    // 获取响应的 Content-Type
    const contentType = response.headers.get('Content-Type');
    // 解析并返回响应数据
    const responseData = await responseFormat(contentType, response);
    // return await response.json();
    if (!configs.notTipBizCodeMsg) {
      handlerBusinessStatus(responseData.code, responseData);
    }
    return responseData;
  } catch (error) {
    // 错误处理
    console.error('Fetch request failed:', error);
    throw error;
  }
}
/**
 * 处理 HTTP 状态码
 * @param {number} status - HTTP 响应状态码
 */
function handleHttpStatus(status) {
  switch (status) {
    case 401:
      console.error('Unauthorized: Authentication is needed or has failed.');
      // 可以在这里添加重新认证的逻辑
      break;
    case 403:
      console.error('Forbidden: The request is understood, but it has been refused.');
      // 可以在这里添加权限不足的逻辑
      break;
    case 404:
      console.error('Not Found: The requested resource could not be found.');
      // 可以在这里添加资源未找到的逻辑
      break;
    case 500:
      console.error('Internal Server Error: An error has occurred on the server side.');
      // 可以在这里添加服务器错误的逻辑
      break;
    case 502:
      console.error('Bad Gateway: The server was acting as a gateway or proxy and received an invalid response from the upstream server.');
      // 可以在这里添加网关错误的逻辑
      break;
    default:
      console.log('Info', `HTTP Status: ${status}`);
      // 可以在这里添加其他状态码的处理逻辑
      break;
  }
}


/**
 * 处理业务状态码的逻辑
 * @param {number} status - response.code 响应状态码
*/
function handlerBusinessStatus(status, responseData) {
  switch (status) {
    case 200:
      break;
    case "000000":
      break;
    case 400:
      break;
    default:
      break;
  }
}

async function responseFormat(contentType, response) {
  let data;
  if (contentType && contentType.includes('application/json')) {
    data = await response.json();
  } else if (contentType && contentType.includes('text')) {
    data = await response.text();
  } else if (contentType && contentType.includes('application/octet-stream')) {
    data = await response.blob();
  } else {
    throw Promise.reject(`Unsupported content type: ${contentType}`);
  }
  return data;
}

function fetchGet(url, params = {}, configs = {}) {
  return fetchRequest('get', url, params, configs);
}
function fetchPost(url, params = {}, configs = {}) {
  return fetchRequest('post', url, params, configs);
}
function fetchPut(url, params = {}, configs = {}) {
  return fetchRequest('put', url, params, configs);
}
function fetchDel(url, params = {}, configs = {}) {
  return fetchRequest('delete', url, params, configs);
}


function requestTimeout(callback, delay = 16) {
  const start = performance.now();
  let handle = null;

  function loop(currentTime) {
    if (currentTime - start >= delay) {
      callback();
    } else {
      handle = requestAnimationFrame(loop);
    }
  }

  handle = requestAnimationFrame(loop);
  return handle;
}

function clearRequestTimeout(handle) {
  cancelAnimationFrame(handle);
}