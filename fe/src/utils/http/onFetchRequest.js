// import { message } from "antd";

/**
 * 封装的 fetchRequest 方法
 * @param {string} method - 请求方法 ("get", "post", "put", "delete")
 * @param {string} url - 请求的 URL
 * @param {Object} params - 请求参数对象
 * @param {Object} configs - 请求函数会用到的配置，例如：headers-请求头对象
 * @returns {Promise} - 返回一个 Promise，解析为响应数据
 * 
 * @example
 * // 示例使用方法
 * fetchRequest('get', 'https://api.example.com/data', { id: 123 }, { headers:{'Authorization': 'Bearer token' }})
 *   .then(data => console.log('Response data:', data))
 *   .catch(error => console.error('Request error:', error));
 * 
 * fetchRequest('post', 'https://api.example.com/data', { name: 'John', age: 30 }, { headers:{'Authorization': 'Bearer token' }})
 *   .then(data => console.log('Response data:', data))
 *   .catch(error => console.error('Request error:', error));
 */
async function fetchRequest(method, url, params = {}, configs = {}) {
  // 初始化请求配置
  const options = {
    method: method.toUpperCase(), // 将方法名转为大写
    headers: {
      'Content-Type': 'application/json', // 默认使用 JSON 形式
      'Pro-Target': globalThis["$ProductionName"] ?? '',
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
  } else if (method.toLowerCase() === 'post' && configs.headers?.urlencoded == 1) { // 表单提交
    const urlParams = new URLSearchParams();
    Object.keys(params?.data).forEach(key => {
      urlParams.append(key, params?.data[key]);
    });
    options.body = urlParams;
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
      globalThis.location.replace('/login');
      break;
    case 403:
      console.error('Forbidden: The request is understood, but it has been refused.');
      // 可以在这里添加权限不足的逻辑
      // globalThis.location.replace('/login');
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
 * 新增的上传资源方法
 * @param {string} url - 上传的 URL
 * @param {File} file - 需要上传的文件
 * @param {Object} params - 附加的请求参数
 * @param {Object} configs - 请求函数会用到的配置
 * @returns {Promise} - 返回一个 Promise，解析为响应数据
 */
async function fetchUploadRequest(method, url, file, params = {}, configs = {}) {
  const formData = new FormData();
  // 检查 file 是否是数组
  if (Array.isArray(file)) {
    file.forEach((f, index) => {
      // 如果是数组，逐个文件添加到 FormData 中，确保字段名唯一
      formData.append(`files`, f);  // 或者 `files[${index}]` 以确保唯一
    });
  } else {
    // 单文件上传，按原方式处理
    formData.append('file', file);   // 将文件添加到 FormData 中
  }
  // formData.append('file', file);

  // 如果有其他参数，添加到 FormData 中
  Object.keys(params).forEach(key => {
    formData.append(key, params[key]);
  });

  // 调用 fetchRequest 来处理上传
  return fetch(url, {
    method: method,  // 如 'POST' 或 'PUT'
    body: formData,  // 使用 FormData 作为请求体
    headers: {
      'Pro-Target': globalThis["$ProductionName"] ?? '',
      // 默认 fetch 自动处理 multipart/form-data 的 Content-Type
      ...configs.headers // 将外部传入的 headers 合并
    },
    ...configs // 合并其他的配置项，例如 mode, credentials 等
  }).then(response => {
    return response.json(); // 解析 JSON 响应
  });
  // return fetchRequest(method, url, formData, {
  //   ...configs,
  //   headers: {
  //     // 不设置 Content-Type，浏览器会自动处理
  //     ...configs.headers,
  //   },
  // });
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
      // message.error(responseData.message);
      break;
    default:
      break;
  }
}

async function responseFormat(contentType, response) {
  let data;
  if (contentType && contentType.includes('application/json')) {
    data = await response.json();
  } else if (contentType && contentType.includes('application/x-www-form-urlencoded')) {
    data = await response.json();
  } else if (contentType && contentType.includes('text')) {
    data = await response.text();
  } else if (contentType && contentType.includes('application/octet-stream')) {
    data = await response.blob();
  } else {
    // throw Promise.reject(`Unsupported content type: ${contentType}`);
  }
  return data;
}

export function fetchGet(url, params = {}, configs = {}) {
  return fetchRequest('get', url, params, configs);
}
export function fetchPost(url, params = {}, configs = {}) {
  return fetchRequest('post', url, params, configs);
}
export function fetchPut(url, params = {}, configs = {}) {
  return fetchRequest('put', url, params, configs);
}
export function fetchDel(url, params = {}, configs = {}) {
  return fetchRequest('delete', url, params, configs);
}
export function fetchUpload(url, file, params = {}, configs = {}) {
  return fetchUploadRequest('post', url, file, params, configs);
}