// 语言类型数字字母映射匹配
export const LANGUAGE_MODE = {
  "01": "cn",
  "02": "en",
  "03": "my",
  "04": "id"
};
// 游戏奖品类型
const GAME_PRICE_LANGUAGE_MODE = {
  "01": "3",
  "02": "5",
  "03": "8"
};

// internationalization 3、8
export const i18n = {
  data: {
    "en": {
      "activeRuleContentTitle": "DESCRIPTION OF GIFT PACKAGE",
      "activeRuleContent": [{
        "text": "This Gift Code is valid from 12/24/2024 to 02/20/2025 (UTC-8). Please redeem it before it expires."
      }, {
        "text": "The Gift Code grants random rewards. Each code can only be redeemed once. Do not share or disclose this page or the Gift Code to others."
      }, {
        "text": "How to redeem: MCGG will be available on the app store starting 12/24/2024. After installing the game, tap your Avatar in the top left to access your Profile, then head to Redemption Code in the Settings located at the top right and enter your Gift Code."
      }, {
        "text"({ lang = "02", mode }) {
          let pathname = location.pathname;
          return `For any questions about the event, <a class="goRulePage" target="_blank" href="${pathname}?lp=1&gpt=11&lang=${lang}&mode=${mode}">check the Event Rules.</a>`
        }
      }]
    },
    "id": {
      "activeRuleContentTitle": "Peraturan Event <br /> Undangan WhatsApp",
      "activeRuleContent": [{
        "text": "1. Kode Hadiah ini berlaku mulai 21/2/2025 - 23/03/2025 (UTC-8). Gunakan sebelum kedaluwarsa."
      }, {
        "text": "2. Tukarkan Kode Hadiah untuk mendapat hadiah acak. Setiap Kode Hadiah hanya bisa ditukarkan 1 kali. Jangan membagikan atau memperlihatkan halaman atau Kode Hadiah ini ke orang lain."
      }, {
        "text": "3. Cara menukar: MCGG akan tersedia di toko aplikasi pada 21/02. Setelah menginstal, ketuk Avatar di sebelah kiri atas untuk membuka Profil, lalu pergi ke Kode Penukaran di Pengaturan yang terletak di kanan atas dan masukkan Kode Hadiah."
      }, {
        "text"({ lang = "04", mode }) {
          let pathname = location.pathname;
          return `4. <a class="goRulePage" target="_blank" href="${pathname}?lp=1&gpt=11&lang=${lang}&mode=${mode}">Lihat Peraturan Event</a> untuk info selengkapnya.`
        }
      }]
    }
  }
}

// 特殊的 1-5 文案
export const i18n_mode1_5 = {
  data: {
    "en": {
      "activeRuleContentTitle": "DESCRIPTION OF GIFT PACKAGE",
      "activeRuleContent": [{
        "text": "This Gift Code is valid from 12/24/2024 to 02/15/2025 (UTC-8). Please redeem it before it expires."
      }, {
        "text": "The Gift Code grants random rewards. Each code can only be redeemed once. Do not share or disclose this page or the Gift Code to others."
      }, {
        "text": "How to redeem: Launch MLBB, tap your Avatar in the top left to access your Profile, then head to Redemption Code in the Settings located at the top right and enter your Gift Code.",
        "imgs": [{
          url: "images/en/content_img_1.png"
        }]
      }, {
        "text"({ lang = "02", mode }) {
          let pathname = location.pathname;
          return `For any questions about the event, <a class="goRulePage" target="_blank" href="${pathname}?lp=1&gpt=11&lang=${lang}&mode=${mode}">check the Event Rules.</a>`
        }
      }]
    },
    "id": {
      "activeRuleContentTitle": "Peraturan Event <br /> Undangan WhatsApp",
      "activeRuleContent": [{
        "text": "1. Kode Hadiah ini berlaku mulai 24/1/2025 - 23/03/2025 (UTC-8). Gunakan sebelum kedaluwarsa."
      }, {
        "text": "2. Tukarkan Kode Hadiah untuk mendapat hadiah acak. Setiap Kode Hadiah hanya bisa ditukarkan 1 kali. Jangan membagikan atau memperlihatkan halaman atau Kode Hadiah ini ke orang lain."
      }, {
        "text": "3. Cara menukar: Buka MLBB, ketuk avatar di kiri atas untuk masuk ke Profil, lalu cari Kode Penukaran pada Pengaturan di kanan atas dan masukkan Kode Hadiah kamu.",
        "imgs": [{
          url: "images/id/content_img_1.png"
        }]
      }, {
        "text"({ lang = "04", mode }) {
          let pathname = location.pathname;
          return `4. Untuk pertanyaan tentang event, <a class="goRulePage" target="_blank" href="${pathname}?lp=1&gpt=11&lang=${lang}&mode=${mode}">harap baca Panduan Event.</a>`
        }
      }]
    }
  }
}


// 获取国际化语言配置数据
export function queryInternationLang(langType = "02", mode) {
  return mode == 1 || mode == 5 ? i18n_mode1_5.data[LANGUAGE_MODE[langType]] : i18n.data[LANGUAGE_MODE[langType]];
}

// 处理国际化数据
export async function handlerInternationalizationTransform(configs = {}) {
  for (let key in configs) {
    let item = configs[key];
    // console.log(key, item, item.activeRuleContent);
    if (!Object.keys(item).length) {
      continue;
    }

    if (!!item?.activeRuleContentTitleImg) {
      let imgUrl = await import(`@assets/${item.activeRuleContentTitleImg}`);
      item.activeRuleContentTitleImg = imgUrl.default;
    }
    if (!!item?.activeRuleWinningInfo) {
      let imgUrl = await import(`@assets/${item.activeRuleWinningInfo}`);
      item.activeRuleWinningInfo = imgUrl.default;
    }
    if (!!item.activeRuleContent) {
      for (let i = 0, len = item.activeRuleContent?.length; i < len; i++) {
        let itemRule = item.activeRuleContent[i];
        if (!itemRule.imgs?.length) {
          itemRule.imgs = [];
        }
        for (let j = 0, len = itemRule.imgs.length; j < len; j++) {
          let _it = itemRule.imgs[j];
          let imgUrl = await import(`@assets/${_it.url}`);
          _it.url = imgUrl.default;
          // console.log('imgUrl', imgUrl, imgUrl.default, _it); // /extensionBundleCode/img/content_img_1.48dda63..png
        }
      }
    }
    // viewData.status.ruleActivity = true;
  }
  return Promise.resolve(configs);
}

// whatsapp message
const whatsppMessage = {
  data: {
    "en": {
      "message"({ code = "" }) {
        return `I'm joining the Magic Chess: Go Go pre-registration event to win a phone, cash, a permanent MLBB Skin, and MCGG Diamonds!\nUse my Code: ${code}`
      }
    },
    "id": {
      "message"({ code = "" }) {
        return `Aku ikutan event Magic Chess: Go Go buat menangin HP, uang tunai, Skin MLBB permanen, dan Diamond MCGG!\nKodeku: ${code}`
      }
    }
  }
}
export function queryWhatsppMessageLang(langType = "02") {
  return whatsppMessage.data[LANGUAGE_MODE[langType]];
}

// 拷贝 CDK 文案
export const copyCDKText = {
  data: {
    "en": {
      title: "【Tersalin】",
      content: [{
        "text": "MCGG will be available on app stores on February 21."
      }, {
        "text": "Please install the game and redeem the code in the game."
      }]
    },
    "id": {
      title: "【Tersalin】",
      content: [{
        "text": "Magic Chess: Go Go akan hadir di toko aplikasi pada 21/02."
      }, {
        "text": "Silakan instal game dan lakukan penukaran di dalam game."
      }]
    }
  }
}
export function queryCopyCDKTextLang(langType = "02") {
  return copyCDKText.data[LANGUAGE_MODE[langType]];
}

// 邀请活动规则
const invitationActivityRules = {
  data: {
    "en": {
      "activeRuleContentTitleImg": "en/rule-tit.png",
      "activeRuleContent": [{
        "text": "This event is organized by Moonton. Please read the event rules and terms carefully before participating. By participating in this event, you acknowledge that you have read, understood, and agreed to all contents of these event rules."
      }, {
        "text": "1. This event is currently available only in Indonesia and is exclusively for WhatsApp users with a country code of (+62)."
      }, {
        "text": "2. Event Period: 01/24/2025 00:00 – 02/20/2025 23:59:59 (UTC-8). The event may end earlier if all rewards are claimed."
      }, {
        "text": "3. Players can join the pre-registration team event via the MCGG WhatsApp business account to earn the following rewards:<br/><br/>Complete Pre-registration: Guaranteed MLBB Lucky Chest containing one of the following: Tigreal \"Lightborn - Defender\", Alucard \"Lightborn - Striker\", Fanny \"Lightborn - Ranger\", Harith \"Lightborn - Inspirer\", or Epic Skin Trial Card Pack (1 Day).<br/><br/>Invite 3 Friends: Guaranteed MCGG Lucky Pack containing one of the following: 1-Star Layla, 1-Star Chou, 1-Star Ling, 1-Star Kagura, or Chess Points ×100, and a chance to win $100cash.<br/><br/>Invite 5 Friends: Guaranteed MLBB Surprise Chest containing one of the following: Wanwan \"Shoujo Commander\", Beatrix \"X Factor\", Granger \"Lightborn - Overrider\", Saber \"Fullmetal Ronin\", Layla \"SABER Destructor\", Hero Fragment ×1, or Ticket ×5, and a chance to win a Vivo phone.<br/><br/>Invite 8 Friends: Guaranteed MCGG Surprise Pack containing one of the following: 1-Star Layla, 1-Star Chou, 1-Star Ling, 1-Star Kagura, Chess Points ×500, Diamonds ×50, Diamonds ×150, Diamonds ×300, and a chance to win $1000 cash."
      }, {
        "text": "Rewards will be distributed via Gift Code. Please keep an eye on WhatsApp notifications.<br/><br/>MLBB: Copy the Gift Code, then open the game, tap your Avatar in the top-left corner, go to Settings in the top-right, and redeem the code.<br/><br/>MCGG: The game will be available on app stores on February 21. After installation, tap your Avatar in the top-left corner, go to Settings in the top-right, and redeem the code."
      }, {
        "text": "Winners will be randomly drawn on 02/27/2025 from eligible participants and contacted via WhatsApp for physical prizes and cash rewards. Winners must provide the required information within 15 days of receiving the notification; otherwise, the prize will be forfeited. If you do not receive a winning notification, it means you did not win. The winner list will be announced on this page on March 1."
      }, {
        "text": "4. Each user can only assist another participant once. Invitations are only considered successful if the invited friend taps the shared link and pre-registers for MCGG via the WhatsApp link."
      }, {
        "text": "5. The organizer reserves the right to interpret and supplement these rules to the maximum extent permitted by law. For any questions, please contact Customer Service in the game's Main Interface."
      }],
      "activeRuleWinningInfo": "en/winning-info-tit.png",
      "activeRuleWinningInfoContent": {
        "columns": [{
          thTitle: "WhatsApp Account"
        }, {
          thTitle: "WhatsApp Name"
        }, {
          thTitle: "Winning Prize"
        }],
        "dataSource": [{
          whatsappAccount: "628578****0381",
          whatsappName: "DIN ****ECT",
          winningPrize: "10 juta rupiah"
        }, {
          whatsappAccount: "628316****5866",
          whatsappName: "Ham***",
          winningPrize: "HP vivo"
        }, {
          whatsappAccount: "628963****7631",
          whatsappName: "デニ***",
          winningPrize: "HP vivo"
        }, {
          whatsappAccount: "628778********10",
          whatsappName: "Fa***",
          winningPrize: "HP vivo"
        }, {
          whatsappAccount: "628535****0992",
          whatsappName: "R****",
          winningPrize: "2 juta rupiah"
        }, {
          whatsappAccount: "628387****9607",
          whatsappName: "RJ.***",
          winningPrize: "2 juta rupiah"
        }, {
          whatsappAccount: "62877********110",
          whatsappName: "F****",
          winningPrize: "2 juta rupiah"
        }, {
          whatsappAccount: "628318****5842",
          whatsappName: "MHD.***",
          winningPrize: "2 juta rupiah"
        }, {
          whatsappAccount: "628589****9453",
          whatsappName: "Ak***",
          winningPrize: "2 juta rupiah"
        }]
      },
    },
    "id": {
      "activeRuleContentTitleImg": "id/rule-tit.png",
      "activeRuleContent": [{
        "text": "Event ini diselenggarakan oleh Moonton. Silakan baca peraturan dan ketentuan event dengan cermat sebelum mengikuti. Dengan mengikuti event ini, kamu mengakui bahwa kamu sudah membaca, memahami, dan menyetujui semua isi dari peraturan event ini."
      }, {
        "text": "1. Saat ini, event hanya terbuka di wilayah Indonesia dan terbatas untuk pengguna WhatsApp dengan kode prefiks (+62)."
      }, {
        "text": "2. Periode Event: 24/01/2025 00:00 - 19/02/2025 23:59:59 (UTC-8). Event bisa berakhir lebih awal setelah hadiah habis."
      }, {
        "text": "3. Pengguna bisa mengikuti event tim praregistrasi melalui akun bisnis WhatsApp MCGG untuk mendapat hadiah yang sesuai.<br/><br/>Menyelesaikan Praregistrasi: Dijamin mendapat MLBB Lucky Chest, buka untuk mendapatkan salah satu item: Tigreal \"Lightborn - Defender\" ×1, Alucard \"Lightborn - Striker\" ×1, Fanny \"Lightborn - Ranger\" ×1, Harith \"Lightborn - Inspirer\" ×1, Epic Skin Trial Pack (1 Hari) ×1<br/><br/>Mengundang 3 Teman: Dijamin mendapat MCGG Lucky Pack, buka untuk mendapatkan salah satu item: Layla Bintang 1 ×1, Chou Bintang 1 ×1, Ling Bintang 1 ×1, Kagura Bintang 1 ×1, Chess Point ×100, dan kesempatan memenangkan uang tunai 2 juta rupiah.<br/><br/>Mengundang 5 Teman: Dijamin mendapat MLBB Surprise Chest, buka untuk mendapatkan salah satu item: Wanwan \"Shoujo Commander\" ×1, Beatrix \"X Factor\" ×1, Granger \"Lightborn\" ×1, Saber \"Fullmetal Ronin\" ×1, Layla \"SABER Destroyer\" ×1, Hero Fragment ×1, Ticket ×5, dan kesempatan memenangkan HP vivo.<br/><br/>Mengundang 8 Teman: Dijamin mendapat MCGG Surprise Pack, buka untuk mendapatkan salah satu item: Layla Bintang 1 ×1, Chou Bintang 1 ×1, Ling Bintang 1 ×1, Kagura Bintang 1 ×1, Chess Point ×500, Diamond ×50, Diamond ×150, Diamond ×300, dan kesempatan memenangkan uang tunai 10 juta rupiah."
      }, {
        "text": "Hadiah dalam game akan dibagikan melalui kode. Silakan cek notifikasi WhatsApp.<br/><br/>MLBB: Salin kode dan login ke game, klik Avatarmu pada kiri atas, lalu tukarkan hadiah pada Pengaturan di kanan atas.<br/><br/>MCGG: MCGG akan hadir di toko aplikasi pada 21/02. Setelah mengunduh game, klik Avatarmu pada kiri atas, lalu tukarkan hadiah pada Pengaturan di kanan atas."
      }, {
        "text": "Hadiah fisik dan hadiah uang tunai akan diundi secara acak dari pengguna yang memenuhi syarat pada 27/02/2025 dan pemenang akan dihubungi melalui WhatsApp. Jika kamu mendapat notifikasi menang, silakan kirim informasi yang dibutuhkan dalam waktu 15 hari, jika tidak, hadiah akan hangus. Jika kamu tidak dapat notifikasi menang, maka artinya kamu tidak menang. Daftar pemenang akan diumumkan di halaman ini pada 1 Maret."
      }, {
        "text": "4. Setiap pengguna hanya bisa membantu 1 kali. Undangan hanya dianggap berhasil ketika teman yang diundang mengetuk link yang dibagikan dan melakukan praregistrasi MCGG melalui link WhatsApp."
      }, {
        "text": "5. Penyelenggara berhak untuk menambahkan dan menafsirkan peraturan event sesuai dengan hukum yang berlaku. Silakan hubungi Customer Service pada interface utama game untuk pertanyaan lebih lanjut."
      }],
      "activeRuleWinningInfo": "id/winning-info-tit.png",
      "activeRuleWinningInfoContent": {
        "columns": [{
          thTitle: "Akaun WhatsApp"
        }, {
          thTitle: "Nama WhatsApp"
        }, {
          thTitle: "Hadiah"
        }],
        "dataSource": [{
          whatsappAccount: "628578****0381",
          whatsappName: "DIN ****ECT",
          winningPrize: "10 juta rupiah"
        }, {
          whatsappAccount: "628316****5866",
          whatsappName: "Ham***",
          winningPrize: "HP vivo"
        }, {
          whatsappAccount: "628963****7631",
          whatsappName: "デニ***",
          winningPrize: "HP vivo"
        }, {
          whatsappAccount: "628778****5110",
          whatsappName: "Fa***",
          winningPrize: "HP vivo"
        }, {
          whatsappAccount: "628535****0992",
          whatsappName: "R****",
          winningPrize: "2 juta rupiah"
        }, {
          whatsappAccount: "628387****9607",
          whatsappName: "RJ.***",
          winningPrize: "2 juta rupiah"
        }, {
          whatsappAccount: "628778****5110",
          whatsappName: "Fa***",
          winningPrize: "2 juta rupiah"
        }, {
          whatsappAccount: "628318****5842",
          whatsappName: "MHD.***",
          winningPrize: "2 juta rupiah"
        }, {
          whatsappAccount: "628589****9453",
          whatsappName: "Ak***",
          winningPrize: "2 juta rupiah"
        }]
      },
    }
  }
}
export function queryActivityRules(langType = "01") {
  // console.log('LANGUAGE_MODE[langType]', LANGUAGE_MODE[langType]);
  return invitationActivityRules.data[LANGUAGE_MODE[langType]];
}