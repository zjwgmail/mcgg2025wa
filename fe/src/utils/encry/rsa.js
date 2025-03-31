import JSEncrypt from 'jsencrypt';

// 加密
export function encrypt(data, publicKeyPem) {
  const encrypt = new JSEncrypt();
  const publicKey = publicKeyPem ?? `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCcbsc7X1y3xn7BvBL/bDCOqfng
ytBvn8mpvgZkOtEMcCLPmZu145BYn01OuZ7HQdb6tK7n7d5/y57avzZyJiAsVGR3
46FaU2AmvoNieoJ96K6GlnKHo8CgAyCwF3dVxp6TfIUHwGs4Z65m73XyXvrbKWW+
BInKK3XoG/qbdxdbpQIDAQAB
-----END PUBLIC KEY-----`;
  encrypt.setPublicKey(publicKey);
  const encryptedData = encrypt.encrypt(data);
  return encryptedData;
}

// 解密
export function decrypt(data, privateKeyPem) {
  const decryptor = new JSEncrypt();
  const privateKey = privateKeyPem ?? `
-----BEGIN RSA PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAJxuxztfXLfGfsG8Ev9sMI6p+eDK0G+fyam+BmQ60QxwIs+Zm7XjkFifTU65nsdB1vq0ruft3n/Lntq/NnImICxUZHfjoVpTYCa+g2J6gn3oroaWcoejwKADILAXd1XGnpN8hQfAazhnrmbvdfJe+tspZb4Eicordegb+pt3F1ulAgMBAAECgYAg7r1oxXG6isJCvPpu5XLvhd9CMNBiv4vv/T5ROYSrDqx1cgwy5Z6M2bSnvzIrFrRQgVtVHmG6G77spFas/1PES+evxGOV5AlXbyck2EwsRIKkIVOkUTAZwUDobF1z9eawDy54W1ko7uRIIDZIMJldSETSWfaKjBs5fwp5jxqb3QJBAOzGq3iVwYEiukyj50NcmKg63M2OEcO21urPTRrePd4zxJG4TrBapB3UT7Px9/InKkPtpdchiEvucdQfuGft3DMCQQCpIjFayOftXNi9YU8aQghYPZ6wiMT6LJOmlWCWjJTZW3bXFbBTqzDaQnYAQzuz9KC98g/Zq++D33TBF6SE2hDHAkEAwF7RZdFWPBL5BdeMx1/t75CTYLZynG5qwq/WV2QFJAkvRa1W0VVzTYD3mJ2Y8zb60eG9AcKOuBJsjQmQi2/nnQJALnycbiR8QqxbUioV0NTHcGF3ZXQiF9T6vDWgd6CqJNfT4Sgv779EzSipQEc6eKrLJ4oJuz1btrZLY+s4p9877wJBAMRM/E56TUPMedcOo7krWi/Rc4jfNWb0FFErNXJO6EEX+LmneUXF+zYqvGWjnC1SxqkYw7rCo+QwHu4lL5CEjMM=
-----END RSA PRIVATE KEY-----
`;
  decryptor.setPrivateKey(privateKey);
  const decryptedData = decryptor.decrypt(data);
  return decryptedData;
}