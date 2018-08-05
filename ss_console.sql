/*
 Navicat Premium Data Transfer

 Source Server         : ofbank
 Source Server Type    : SQLite
 Source Server Version : 3012001
 Source Schema         : main

 Target Server Type    : SQLite
 Target Server Version : 3012001
 File Encoding         : 65001

 Date: 05/08/2018 14:36:46
*/

PRAGMA foreign_keys = false;

-- ----------------------------
-- Table structure for ss_client_config
-- ----------------------------
DROP TABLE IF EXISTS "ss_client_config";
CREATE TABLE "ss_client_config" (
  "id" integer PRIMARY KEY AUTOINCREMENT,
  "c_port" integer,
  "c_password" VARCHAR(50,2),
  "create_time" DATETIME DEFAULT (datetime(CURRENT_TIMESTAMP,'localtime'))
);

-- ----------------------------
-- Table structure for ss_server_config
-- ----------------------------
DROP TABLE IF EXISTS "ss_server_config";
CREATE TABLE "ss_server_config" (
  "id" integer PRIMARY KEY AUTOINCREMENT,
  "server_ip" VARCHAR(50,2),
  "encrypt_way" VARCHAR(50,2),
  "create_time" DATETIME DEFAULT (datetime(CURRENT_TIMESTAMP,'localtime'))
);

-- ----------------------------
-- Auto increment value for ss_client_config
-- ----------------------------

-- ----------------------------
-- Auto increment value for ss_server_config
-- ----------------------------

PRAGMA foreign_keys = true;
