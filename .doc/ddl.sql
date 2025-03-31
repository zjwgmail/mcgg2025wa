-- `wa-fission`.activity_info definition

CREATE TABLE `activity_info`
(
    `id`              int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `activity_name`   varchar(256) NOT NULL DEFAULT '' COMMENT '活动名称',
    `activity_status` varchar(54)           DEFAULT 'unstart' COMMENT '活动状态：unstart：未开始；started:已开始；buffer:缓冲期；end：结束',
    `created_at`      datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`      datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `start_at`        datetime     NOT NULL COMMENT '活动开始时间',
    `end_at`          datetime     NOT NULL COMMENT '结束时间（开始-结束-缓冲）',
    `end_buffer_day`  int(11) DEFAULT NULL COMMENT '结束时间后的缓冲天数',
    `end_buffer_at`   datetime              DEFAULT NULL COMMENT '进入缓冲期的时间',
    `really_end_at`   datetime              DEFAULT NULL COMMENT '真正结束的时间',
    `help_max`        int(11) NOT NULL COMMENT '最大助力次数',
    `cost_max` double(64,2) DEFAULT NULL COMMENT '活动预算上限',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='活动表';
INSERT INTO `wa-fission`.activity_info
(id, activity_name, start_at, end_at, end_buffer_day, end_buffer_at,
 really_end_at, help_max)
VALUES (1, 'fission', '2024-12-01 00:00:00',
        '2025-12-31 23:59:59', 1, NULL, NULL, 8);
-- `wa-fission`.cost_count_info definition

CREATE TABLE `cost_count_info`
(
    `id`         int(11) NOT NULL COMMENT '活动id',
    `cost_count` double(64,4) unsigned NOT NULL DEFAULT '0.0000' COMMENT '消费总计',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='费用统计表';
-- `wa-fission`.free_cdk_info definition

CREATE TABLE `free_cdk_info`
(
    `id`          int(11) NOT NULL AUTO_INCREMENT,
    `del`         tinyint(1) NOT NULL DEFAULT '0',
    `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
    `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `wa_id`       varchar(255) NOT NULL,
    `create_at`   bigint(20) NOT NULL,
    `send_state`  tinyint(4) DEFAULT '1' COMMENT '发送状态 1:未发送 2:已发送',
    `send_at`     bigint(20) DEFAULT NULL COMMENT '发送时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `wa_uniq` (`wa_id`) USING BTREE,
    KEY           `create_idx` (`create_at`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='免费CDK信息';
-- `wa-fission`.help_info_v2 definition

CREATE TABLE `help_info_v2`
(
    `id`          int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `rally_code`  varchar(256) NOT NULL DEFAULT '' COMMENT '集结码（被助力人的集结码）',
    `wa_id`       varchar(256) NOT NULL DEFAULT '' COMMENT '助力人whatsappid',
    `created_at`  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `help_status` varchar(255) NOT NULL COMMENT '助力状态（待定）：efficien：生效；unefficien：未生效',
    `help_at`     bigint(20) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `joiner_uniq` (`wa_id`) USING BTREE,
    KEY           `code_idx` (`rally_code`) USING BTREE,
    KEY           `help_at_idx` (`help_at`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='助力表';
-- `wa-fission`.msg_info_v2 definition

CREATE TABLE `msg_info_v2`
(
    `id`               varchar(256) NOT NULL COMMENT '雪花id',
    `type`             varchar(32)  NOT NULL DEFAULT '' COMMENT 'send：发送；receive：接收',
    `msg`              text         NOT NULL COMMENT '消息内容',
    `msg_status`       varchar(64)  NOT NULL DEFAULT 'owner_un_send' COMMENT 'owner_un_send:后台未发送；owner_send:后台已发送；send：牛信云已发送；failed：牛信云发送失败',
    `wa_id`            varchar(256) NOT NULL DEFAULT '' COMMENT '与商户号交互的whatsappid',
    `msg_type`         varchar(255) NOT NULL COMMENT '开团消息；红包召回；红包领取消息，奖励领取，进度更新',
    `currency`         varchar(256)          DEFAULT '' COMMENT '币种',
    `price` double(64,4) unsigned DEFAULT '0.0000' COMMENT '客户售价,本币CNY',
    `foreign_price` double(64,4) unsigned DEFAULT '0.0000' COMMENT '客户售价,外币',
    `wa_message_id`    varchar(256)          DEFAULT '' COMMENT 'wa消息id',
    `created_at`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `msg_at`           bigint(20) NOT NULL,
    `is_count`         tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否已统计',
    `source_wa_id`     varchar(256)          DEFAULT '' COMMENT '由此whatsappid导致的发消息',
    `receive_msg`      text COMMENT '回调的消息',
    `trace_id`         varchar(256)          DEFAULT NULL COMMENT '链路id',
    `send_res`         text COMMENT '发送返回结果',
    `build_msg_params` text COMMENT '构建消息的参数',
    PRIMARY KEY (`id`),
    KEY                `idx_wa_msg_id` (`wa_message_id`) USING BTREE,
    KEY                `user_idx` (`wa_id`) USING BTREE,
    KEY                `msg_at_idx` (`msg_at`) USING BTREE,
    KEY                `foreign_price_idx` (`foreign_price`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表';
-- `wa-fission`.report_msg_info definition

CREATE TABLE `report_msg_info`
(
    `id`          varchar(256) NOT NULL COMMENT '雪花id',
    `date`        varchar(256) NOT NULL COMMENT '报告日期',
    `hour`        varchar(256)          DEFAULT '' COMMENT '报告时间',
    `report_type` varchar(54)  NOT NULL DEFAULT 'feishu' COMMENT '报告类型：feishu：飞书；excel:excel表格；',
    `msg_status`  varchar(64)  NOT NULL DEFAULT 'owner_un_send' COMMENT 'owner_un_send:后台未发送；fail：发送失败；owner_send:后台已发送；',
    `msg`         text         NOT NULL COMMENT '发送消息内容',
    `count_msg`   text,
    `res`         varchar(1024)         DEFAULT NULL COMMENT '发送结果',
    `created_at`  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='报告消息表';
-- `wa-fission`.rsv_msg_info definition

CREATE TABLE `rsv_msg_info`
(
    `id`               varchar(256) NOT NULL COMMENT '雪花id',
    `type`             varchar(32)  NOT NULL DEFAULT '' COMMENT 'send：发送；receive：接收',
    `msg`              text         NOT NULL COMMENT '消息内容',
    `msg_status`       varchar(64)  NOT NULL DEFAULT 'owner_un_send' COMMENT 'owner_un_send:后台未发送；owner_send:后台已发送；send：牛信云已发送；failed：牛信云发送失败',
    `wa_id`            varchar(256) NOT NULL DEFAULT '' COMMENT '与商户号交互的whatsappid',
    `msg_type`         varchar(255) NOT NULL COMMENT '开团消息；红包召回；红包领取消息，奖励领取，进度更新',
    `currency`         varchar(256)          DEFAULT '' COMMENT '币种',
    `price` double(64,4) unsigned DEFAULT '0.0000' COMMENT '客户售价,本币CNY',
    `foreign_price` double(64,4) unsigned DEFAULT '0.0000' COMMENT '客户售价,外币',
    `wa_message_id`    varchar(256)          DEFAULT '' COMMENT 'wa消息id',
    `created_at`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_count`         tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否已统计',
    `source_wa_id`     varchar(256)          DEFAULT '' COMMENT '由此whatsappid导致的发消息',
    `receive_msg`      text COMMENT '回调的消息',
    `trace_id`         varchar(256)          DEFAULT NULL COMMENT '链路id',
    `send_res`         text COMMENT '发送返回结果',
    `build_msg_params` text COMMENT '构建消息的参数',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='接收消息表';
-- `wa-fission`.user_attend_info_v2 definition

CREATE TABLE `user_attend_info_v2`
(
    `id`                         int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `channel`                    varchar(10)           DEFAULT '' COMMENT '渠道',
    `language`                   varchar(10)           DEFAULT '' COMMENT '语言',
    `generation`                 varchar(10)           DEFAULT '' COMMENT '代数',
    `identification_code`        varchar(50)           DEFAULT '' COMMENT '玩家识别码',
    `wa_id`                      varchar(256) NOT NULL DEFAULT '' COMMENT 'whatsappid',
    `rally_code`                 varchar(100)          DEFAULT '' COMMENT '集结码',
    `user_nickname`              varchar(256)          DEFAULT '' COMMENT '用户昵称',
    `three_cdk_code`             varchar(256)          DEFAULT '' COMMENT '三人cdk',
    `five_cdk_code`              varchar(256)          DEFAULT '' COMMENT '五人cdk',
    `eight_cdk_code`             varchar(256)          DEFAULT '' COMMENT '八人cdk',
    `attend_at`                  bigint(20) NOT NULL COMMENT '参与时间',
    `start_group_at`             datetime              DEFAULT NULL COMMENT '开团时间',
    `newest_free_start_at`       bigint(20) NOT NULL COMMENT '最近免费开始时间（最近一次用户主动给商户号发送消息的时间）',
    `newest_free_end_at`         datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最近免费结束时间',
    `send_renew_free_at`         bigint(20) DEFAULT NULL COMMENT '最近免费结束时间',
    `is_send_renew_free_msg`     tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否发送了续免费消息',
    `newest_help_at`             datetime              DEFAULT NULL COMMENT '最近助力时间(初始值是开团时间)',
    `three_over_at`              datetime              DEFAULT NULL COMMENT '三人助力成功时间',
    `five_over_at`               datetime              DEFAULT NULL COMMENT '五人助力成功时间',
    `eight_over_at`              datetime              DEFAULT NULL COMMENT '八人助力成功时间',
    `attend_status`              varchar(32)  NOT NULL DEFAULT 'attend' COMMENT 'attend:参与活动；start_group：开团；three_over:三人助力成功；five_over:五人助力成功；eight_over:八人助力成功',
    `is_three_stage`             tinyint(4) DEFAULT '1' COMMENT '1:没达到三人；2达到三人',
    `is_five_stage`              tinyint(4) DEFAULT '1' COMMENT '1:达到五人 ；2:未达到五人',
    `created_at`                 datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`                 datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `extra`                      text COMMENT '用户参与消息内容',
    `is_send_cdk_msg`            tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否发送了cdk消息',
    `is_send_clustering_msg`     tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否发送了催促成团消息',
    `send_clustering_at`         bigint(20) DEFAULT NULL COMMENT '发送催促成团消息的时间',
    `is_send_pay_renew_free_msg` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否发送了付费的续免费消息',
    `short_link`                 varchar(256)          DEFAULT NULL,
    `has_helper`                 tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否有助力人 0否 1是',
    PRIMARY KEY (`id`),
    UNIQUE KEY `user_uniq` (`wa_id`) USING BTREE,
    KEY                          `code_index` (`rally_code`) USING BTREE,
    KEY                          `attend_at_index` (`attend_at`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户参与信息表';
-- `wa-fission`.rsv_other_msg_info_1 definition

CREATE TABLE `rsv_other_msg_info_1`
(
    `id`         int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `msg`        text         NOT NULL COMMENT '消息内容',
    `wa_id`      varchar(256) NOT NULL DEFAULT '' COMMENT '与商户号交互的whatsappid',
    `timestamp`  bigint(32) DEFAULT '0' COMMENT '消息时间戳',
    `created_at` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY          `idx_timestamp_waid` (`timestamp`,`wa_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='接收其他系统消息表';