import CryptoJS from 'crypto-js'
const keyStr = 'w;skfjnxm.?s23=1'

export default {
  // 加密
  encrypt (word) {
    const key = CryptoJS.enc.Utf8.parse(keyStr)
    const srcs = CryptoJS.enc.Utf8.parse(word)
    const encrypted = CryptoJS.AES.encrypt(srcs, key, { mode: CryptoJS.mode.ECB, padding: CryptoJS.pad.Pkcs7 })
    return encrypted.toString()
  },
  // 解密
  decrypt (word) {
    const key = CryptoJS.enc.Utf8.parse(keyStr)
    const decrypt = CryptoJS.AES.decrypt(word, key, { mode: CryptoJS.mode.ECB, padding: CryptoJS.pad.Pkcs7 })
    return CryptoJS.enc.Utf8.stringify(decrypt).toString()
  }
}
