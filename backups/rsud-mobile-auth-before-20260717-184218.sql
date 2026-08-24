-- MySQL dump 10.13  Distrib 8.4.9, for Linux (x86_64)
--
-- Host: localhost    Database: rsud_otista
-- ------------------------------------------------------
-- Server version	8.4.9

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `user_mobile`
--

DROP TABLE IF EXISTS `user_mobile`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_mobile` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `username` varchar(100) NOT NULL,
  `email` varchar(100) NOT NULL,
  `no_rm` varchar(20) DEFAULT NULL,
  `patient_id` bigint unsigned DEFAULT NULL,
  `phone` varchar(30) DEFAULT NULL,
  `full_name` varchar(150) DEFAULT NULL,
  `password` varchar(255) NOT NULL,
  `is_deleted` tinyint(1) DEFAULT '0',
  `deleted_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `email_verified` tinyint(1) DEFAULT '0',
  `verification_token` varchar(255) DEFAULT NULL,
  `verified_at` timestamp NULL DEFAULT NULL,
  `medical_record_verified_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `uk_user_patient_id` (`patient_id`),
  KEY `idx_user_mobile_no_rm` (`no_rm`)
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_mobile`
--

LOCK TABLES `user_mobile` WRITE;
/*!40000 ALTER TABLE `user_mobile` DISABLE KEYS */;
INSERT INTO `user_mobile` VALUES (1,'fery','fery@gmail.com',NULL,NULL,NULL,NULL,'$2a$10$L20f4JCmwdd86JYaZ2Art.eDsKgU1OfE6y3G9l78unI8JXgvVHqLC',0,NULL,'2026-05-19 05:37:35','2026-05-19 05:37:35',0,NULL,NULL,NULL),(3,'fery2','fery2@gmail.com',NULL,NULL,NULL,NULL,'$2a$10$3bMM/r4I0XBzdgq45bm4MOHc8Otei0vp.9ivAtiYDEd4ktgpA0zCO',0,NULL,'2026-05-19 07:26:21','2026-05-20 07:26:24',1,NULL,NULL,NULL),(4,'fery23','asdasd@gmail.com',NULL,NULL,NULL,NULL,'$2a$10$v8Mu7hs2Cp2lUPz6ES47GOdV6AUQ2CFFoYA5taDjH8O9d1qkx82IW',0,NULL,'2026-05-19 07:28:59','2026-05-20 14:59:11',1,'53b0112e-0bb4-45eb-af19-ec627e3211c0',NULL,NULL),(5,'tes','VeblenismSecaucus@harakirimail.com',NULL,NULL,NULL,NULL,'$2a$10$H59rAAP/00Zz959OSacMyu5ae6VodEsVrzPoEeLuNNfjhkV3nVtSG',0,NULL,'2026-05-19 07:31:32','2026-05-20 10:08:03',1,'57ea2725-deac-44ec-8d7e-b52f948d36b2',NULL,NULL),(6,'fery12345','asd',NULL,NULL,NULL,NULL,'$2a$10$jih44bLFywTWdKFaDh2K1OHnD.g8V4x47pH9KxcT/ztocHfuTf74i',0,NULL,'2026-05-21 02:09:33','2026-05-21 02:09:56',0,'',NULL,NULL),(7,'asd','asd123',NULL,NULL,NULL,NULL,'$2a$10$YIHMCa311yI8PfIw/1DrHu2OTT81D35G0iBCgsuvjCkGQJOWtsklG',0,NULL,'2026-05-21 02:10:10','2026-05-21 02:11:31',0,'',NULL,NULL),(8,'1231231','23421341',NULL,NULL,NULL,NULL,'$2a$10$VmM3Ydioo7GO8ukrhLoYb.1wEGSPkSlUoD96kVDw/mCuzeZHijqLG',0,NULL,'2026-05-21 02:11:35','2026-05-21 02:12:02',0,'',NULL,NULL),(9,'fery12345213','aji.ferybayu@gmail.com',NULL,NULL,NULL,NULL,'$2a$10$Sl6nWA46Kt5EPw267fi/SuLqV5vNRfcPUhWFsOm7Cww397cAjWfBa',0,NULL,'2026-05-21 02:14:16','2026-05-21 02:14:16',1,'',NULL,NULL),(10,'Maulana Ganda Wijaya','maulanagandawijaya@gmail.com','810739',766594,'08999228241','Maulana Ganda Wijaya','$2a$10$IkM.9C7VtEP4l39CriOiWOwmC4vC7zb6KX2rFkBDTrVUsrDp3cWRm',0,NULL,'2026-05-24 05:47:14','2026-06-08 17:25:16',1,'',NULL,'2026-06-08 17:25:16'),(11,'oki anandari','oki.anandari.h@gmail.com','651379',557276,'082216168491','oki anandari','$2a$10$jfzmGWeO5AsJ2u4IwMysV.amZXZFL/StLkf1W11TX/Sr5bnSSrcsW',0,NULL,'2026-05-30 04:37:50','2026-06-25 06:35:19',1,NULL,NULL,'2026-06-25 06:35:19'),(12,'test','necrology69@gmail.com','160136',657538,'08999228241','test','$2a$10$3OfboohL1hkLmzFBeKymruNvbwl170XGzvFsRhQN0vggYhBH3l09K',0,NULL,'2026-06-11 14:23:56','2026-06-23 14:46:11',1,NULL,NULL,'2026-06-11 14:26:16'),(13,'aditya wicaksono','diditaditya35@gmail.com','123456',657292,'081313616045','aditya wicaksono','$2a$10$xmOPqC2/tnOZ4QGNJdjeY.uhhQQ3tUThDpbjHSfNj28m/OSWnSVjG',0,NULL,'2026-06-15 01:18:53','2026-06-22 04:25:11',1,NULL,NULL,'2026-06-22 04:25:11'),(14,'setyo','setyo_witjaksono@yahoo.com','260838',659575,'081380030570','setyo','$2a$10$e0o5SWuR81LOjj6.3bLAHehy9HsVqf5KC6xe3xVMlb4IcXIj5U55e',0,NULL,'2026-06-26 05:41:26','2026-06-26 05:48:06',1,NULL,NULL,'2026-06-26 05:48:06'),(15,'cecep budi amrizal','towermasjid@gmail.com','840123',796035,'081313494030','cecep budi amrizal','$2a$10$fG/TzW/MbDjJxizSyQ1/xuWCh6VapKcKMo1.hN/1ky2SAA9GowR.u',0,NULL,'2026-06-26 05:53:15','2026-06-26 05:55:22',1,NULL,NULL,'2026-06-26 05:55:22');
/*!40000 ALTER TABLE `user_mobile` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `otp_user_mobile`
--

DROP TABLE IF EXISTS `otp_user_mobile`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `otp_user_mobile` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `otp_code` varchar(10) NOT NULL,
  `expired_at` timestamp NOT NULL,
  `is_used` tinyint(1) DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_otp_user_mobile_user_used_expired` (`user_id`,`is_used`,`expired_at`),
  CONSTRAINT `otp_user_mobile_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `user_mobile` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=36 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `otp_user_mobile`
--

LOCK TABLES `otp_user_mobile` WRITE;
/*!40000 ALTER TABLE `otp_user_mobile` DISABLE KEYS */;
INSERT INTO `otp_user_mobile` VALUES (1,5,'098419','2026-05-20 08:15:41',0,'2026-05-20 08:10:41'),(2,5,'387683','2026-05-20 08:16:06',0,'2026-05-20 08:11:06'),(3,5,'150403','2026-05-20 08:19:22',0,'2026-05-20 08:14:22'),(4,5,'361577','2026-05-20 09:43:08',0,'2026-05-20 09:38:08'),(5,4,'992105','2026-05-20 09:49:37',0,'2026-05-20 09:44:37'),(6,4,'506042','2026-05-20 09:49:42',0,'2026-05-20 09:44:42'),(7,4,'188908','2026-05-20 09:49:49',0,'2026-05-20 09:44:49'),(8,4,'289049','2026-05-20 09:49:54',0,'2026-05-20 09:44:54'),(9,4,'429982','2026-05-20 09:50:50',0,'2026-05-20 09:45:50'),(10,4,'384718','2026-05-20 09:52:56',0,'2026-05-20 09:47:56'),(11,4,'948510','2026-05-20 09:56:08',0,'2026-05-20 09:51:08'),(12,4,'559634','2026-05-20 10:03:02',1,'2026-05-20 09:58:02'),(13,4,'718924','2026-05-20 10:07:37',0,'2026-05-20 10:02:37'),(14,4,'479452','2026-05-20 10:07:44',0,'2026-05-20 10:02:44'),(15,4,'437183','2026-05-20 10:09:41',0,'2026-05-20 10:04:41'),(16,5,'091851','2026-05-20 10:12:01',0,'2026-05-20 10:07:01'),(17,5,'185803','2026-05-20 10:13:11',0,'2026-05-20 10:08:11'),(18,4,'283486','2026-05-20 10:13:44',0,'2026-05-20 10:08:44'),(19,4,'845229','2026-05-20 12:23:58',0,'2026-05-20 12:18:58'),(20,10,'282362','2026-05-24 05:52:52',1,'2026-05-24 05:47:52'),(21,10,'913187','2026-05-24 15:00:20',1,'2026-05-24 14:55:20'),(22,10,'980393','2026-05-24 15:02:38',1,'2026-05-24 14:57:38'),(23,10,'923413','2026-05-24 15:11:07',1,'2026-05-24 15:06:07'),(24,10,'953537','2026-05-29 14:10:42',1,'2026-05-29 14:05:42'),(25,10,'072840','2026-05-30 04:22:26',1,'2026-05-30 04:17:26'),(26,10,'099972','2026-06-08 12:51:15',1,'2026-06-08 12:46:15'),(27,10,'248011','2026-06-10 15:21:27',0,'2026-06-10 15:16:27'),(28,10,'687687','2026-06-10 15:26:06',1,'2026-06-10 15:21:06'),(29,12,'911882','2026-06-11 15:01:16',1,'2026-06-11 14:56:16'),(30,12,'374928','2026-06-13 00:22:41',1,'2026-06-13 00:17:41'),(31,13,'089460','2026-06-22 04:27:28',1,'2026-06-22 04:22:28'),(32,10,'874259','2026-06-23 14:48:38',1,'2026-06-23 14:43:38'),(33,12,'430382','2026-06-23 14:51:19',0,'2026-06-23 14:46:19'),(34,12,'037964','2026-06-23 14:51:31',1,'2026-06-23 14:46:31'),(35,11,'095812','2026-06-25 06:39:16',1,'2026-06-25 06:34:16');
/*!40000 ALTER TABLE `otp_user_mobile` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `otp_verif_email_mobile`
--

DROP TABLE IF EXISTS `otp_verif_email_mobile`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `otp_verif_email_mobile` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `username` varchar(100) NOT NULL,
  `email` varchar(100) NOT NULL,
  `no_rm` varchar(20) DEFAULT NULL,
  `phone` varchar(30) DEFAULT NULL,
  `full_name` varchar(150) DEFAULT NULL,
  `otp_code` varchar(10) NOT NULL,
  `expired_at` timestamp NOT NULL,
  `is_used` tinyint(1) DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_otp_verif_email_mobile_email_used_expired` (`email`,`is_used`,`expired_at`)
) ENGINE=InnoDB AUTO_INCREMENT=20 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `otp_verif_email_mobile`
--

LOCK TABLES `otp_verif_email_mobile` WRITE;
/*!40000 ALTER TABLE `otp_verif_email_mobile` DISABLE KEYS */;
INSERT INTO `otp_verif_email_mobile` VALUES (1,'fery234','asdasd',NULL,NULL,NULL,'602924','2026-05-20 12:25:15',0,'2026-05-20 12:20:15'),(2,'fery1234','asdasd',NULL,NULL,NULL,'199047','2026-05-20 14:51:39',0,'2026-05-20 14:46:39'),(3,'fery1234','as',NULL,NULL,NULL,'834011','2026-05-20 14:52:05',0,'2026-05-20 14:47:05'),(4,'fery1234','asda',NULL,NULL,NULL,'815626','2026-05-20 14:54:37',1,'2026-05-20 14:49:37'),(5,'fery1234','asdasd',NULL,NULL,NULL,'127652','2026-05-20 14:58:14',0,'2026-05-20 14:53:14'),(6,'fery1234','asdasdas',NULL,NULL,NULL,'929585','2026-05-20 14:59:12',0,'2026-05-20 14:54:12'),(7,'fery1234','asfasd',NULL,NULL,NULL,'147866','2026-05-20 15:02:09',0,'2026-05-20 14:57:09'),(8,'fery1234','asd',NULL,NULL,NULL,'252610','2026-05-20 15:05:38',0,'2026-05-20 15:00:38'),(9,'fery1234','asd',NULL,NULL,NULL,'248492','2026-05-20 15:06:00',0,'2026-05-20 15:01:00'),(10,'fery12345','asd',NULL,NULL,NULL,'466742','2026-05-20 15:09:54',0,'2026-05-20 15:04:54'),(11,'fery12345','123',NULL,NULL,NULL,'927461','2026-05-20 15:10:51',0,'2026-05-20 15:05:51'),(12,'fery12345213','aji.ferybayu@gmail.com',NULL,NULL,NULL,'227417','2026-05-20 15:12:40',1,'2026-05-20 15:07:40'),(13,'Maulana Ganda Wijaya','maulanagandawijaya@gmail.com',NULL,'08999228241','Maulana Ganda Wijaya','521885','2026-05-24 05:51:44',1,'2026-05-24 05:46:44'),(14,'oki anandari','oki.anandari.h@gmail.com',NULL,'082216168491','oki anandari','363117','2026-05-30 04:42:10',1,'2026-05-30 04:37:10'),(15,'test','necrology69@gmail.com',NULL,'08999228241','test','903781','2026-06-11 14:27:51',1,'2026-06-11 14:22:51'),(16,'aditya','diditaditya35@gmail.com',NULL,'0813136161044','aditya','124153','2026-06-13 01:03:52',0,'2026-06-13 00:58:52'),(17,'aditya wicaksono','diditaditya35@gmail.com',NULL,'081313616045','aditya wicaksono','682257','2026-06-15 01:23:24',1,'2026-06-15 01:18:24'),(18,'setyo','setyo_witjaksono@yahoo.com',NULL,'081380030570','setyo','715396','2026-06-26 05:45:34',1,'2026-06-26 05:40:34'),(19,'cecep budi amrizal','towermasjid@gmail.com',NULL,'081313494030','cecep budi amrizal','963902','2026-06-26 05:53:48',1,'2026-06-26 05:48:48');
/*!40000 ALTER TABLE `otp_verif_email_mobile` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `otp_password_reset_mobile`
--

DROP TABLE IF EXISTS `otp_password_reset_mobile`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `otp_password_reset_mobile` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `otp_code` varchar(10) NOT NULL,
  `expired_at` timestamp NOT NULL,
  `is_used` tinyint(1) DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_otp_password_reset_mobile_user_used_expired` (`user_id`,`is_used`,`expired_at`),
  CONSTRAINT `otp_password_reset_mobile_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `user_mobile` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `otp_password_reset_mobile`
--

LOCK TABLES `otp_password_reset_mobile` WRITE;
/*!40000 ALTER TABLE `otp_password_reset_mobile` DISABLE KEYS */;
INSERT INTO `otp_password_reset_mobile` VALUES (1,10,'937621','2026-05-24 15:01:08',0,'2026-05-24 14:56:08'),(2,10,'802357','2026-05-24 15:10:31',1,'2026-05-24 15:05:31'),(3,12,'883972','2026-06-23 14:50:10',1,'2026-06-23 14:45:10'),(4,11,'052786','2026-06-25 06:37:32',1,'2026-06-25 06:32:32'),(5,11,'826852','2026-06-25 06:38:46',1,'2026-06-25 06:33:46');
/*!40000 ALTER TABLE `otp_password_reset_mobile` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `otp_medical_record_claim_mobile`
--

DROP TABLE IF EXISTS `otp_medical_record_claim_mobile`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `otp_medical_record_claim_mobile` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `patient_id` bigint unsigned NOT NULL,
  `no_rm` varchar(20) NOT NULL,
  `patient_name` varchar(150) DEFAULT NULL,
  `otp_code` varchar(10) NOT NULL,
  `expired_at` timestamp NOT NULL,
  `is_used` tinyint(1) DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_otp_medical_record_claim_user` (`user_id`),
  KEY `idx_otp_medical_record_claim_no_rm` (`no_rm`),
  KEY `idx_otp_medical_record_claim_mobile_user_used_expired` (`user_id`,`is_used`,`expired_at`),
  CONSTRAINT `otp_medical_record_claim_mobile_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `user_mobile` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `otp_medical_record_claim_mobile`
--

LOCK TABLES `otp_medical_record_claim_mobile` WRITE;
/*!40000 ALTER TABLE `otp_medical_record_claim_mobile` DISABLE KEYS */;
INSERT INTO `otp_medical_record_claim_mobile` VALUES (1,10,766594,'810739','COBA TEST','860237','2026-06-08 17:29:28',1,'2026-06-08 17:24:28'),(2,12,657538,'160136','FAHRA AULIA','853494','2026-06-11 14:30:09',0,'2026-06-11 14:25:09'),(3,12,657538,'160136','FAHRA AULIA','195846','2026-06-11 14:30:37',1,'2026-06-11 14:25:37'),(4,13,657292,'123456','ADI SUWARDANI','573111','2026-06-22 04:29:40',1,'2026-06-22 04:24:40'),(5,11,557276,'651379','OKI ANANDARI HIDAYAT, ST','884364','2026-06-25 06:40:04',1,'2026-06-25 06:35:04'),(6,14,659575,'260838','SETYO WITJAKSONO','291005','2026-06-26 05:52:22',1,'2026-06-26 05:47:22'),(7,15,796035,'840123','CECEP BUDI AMRIZAL','266287','2026-06-26 05:59:56',1,'2026-06-26 05:54:56');
/*!40000 ALTER TABLE `otp_medical_record_claim_mobile` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Dumping events for database 'rsud_otista'
--

--
-- Dumping routines for database 'rsud_otista'
--
/*!50003 DROP PROCEDURE IF EXISTS `laporan_rm_profil_54` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`oki`@`%` PROCEDURE `laporan_rm_profil_54`(
    IN p_tahun INT
)
BEGIN

    DECLARE v_tanggal_awal DATETIME;
    DECLARE v_tanggal_akhir DATETIME;

    SET v_tanggal_awal = CONCAT(p_tahun, '-01-01 00:00:00');
    SET v_tanggal_akhir = CONCAT(p_tahun, '-12-31 23:59:59');

    SELECT
        1 AS no,
        'RSUD SOREANG' AS jenis_pelayanan,

        -- =====================================================
        -- RAWAT JALAN
        -- =====================================================

        SUM(
            CASE
                WHEN poli.id != '19'
                AND p.kelamin = 'L'
                THEN 1 ELSE 0
            END
        ) AS irj_l,

        SUM(
            CASE
                WHEN poli.id != '19'
                AND p.kelamin = 'P'
                THEN 1 ELSE 0
            END
        ) AS irj_p,

        SUM(
            CASE
                WHEN poli.id != '19'
                THEN 1 ELSE 0
            END
        ) AS irj_total,

        -- =====================================================
        -- GANGGUAN JIWA
        -- =====================================================

        SUM(
            CASE
                WHEN poli.id = '19'
                AND p.kelamin = 'L'
                THEN 1 ELSE 0
            END
        ) AS jiwa_l,

        SUM(
            CASE
                WHEN poli.id = '19'
                AND p.kelamin = 'P'
                THEN 1 ELSE 0
            END
        ) AS jiwa_p,

        SUM(
            CASE
                WHEN poli.id = '19'
                THEN 1 ELSE 0
            END
        ) AS jiwa_total,

        -- =====================================================
        -- RAWAT INAP
        -- =====================================================

        (
            SELECT COUNT(*)
            FROM rawatinaps ri
						JOIN registrasis r ON r.id = ri.registrasi_id
            JOIN pasiens px ON px.id = r.pasien_id
            WHERE ri.created_at >= v_tanggal_awal
              AND ri.created_at <= v_tanggal_akhir
              AND ri.deleted_at IS NULL
              AND px.kelamin = 'L'
        ) AS iri_l,

        (
            SELECT COUNT(*)
            FROM rawatinaps ri
						JOIN registrasis r ON r.id = ri.registrasi_id
            JOIN pasiens px ON px.id = r.pasien_id
            WHERE ri.created_at >= v_tanggal_awal
              AND ri.created_at <= v_tanggal_akhir
              AND ri.deleted_at IS NULL
              AND px.kelamin = 'P'
        ) AS iri_p,

        (
            SELECT COUNT(*)
            FROM rawatinaps ri
						JOIN registrasis r ON r.id = ri.registrasi_id
            WHERE ri.created_at >= v_tanggal_awal
              AND ri.created_at <= v_tanggal_akhir
              AND ri.deleted_at IS NULL
        ) AS iri_total

    FROM registrasis r
    JOIN pasiens p ON p.id = r.pasien_id
    LEFT JOIN polis poli ON poli.id = r.poli_id

    WHERE r.created_at >= v_tanggal_awal
      AND r.created_at <= v_tanggal_akhir
      AND r.deleted_at IS NULL;

END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `laporan_rm_profil_55` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `laporan_rm_profil_55`(
    IN p_tahun INT
)
BEGIN

    DECLARE v_tanggal_awal DATETIME;
    DECLARE v_tanggal_akhir DATETIME;

    SET v_tanggal_awal = CONCAT(p_tahun, '-01-01 00:00:00');
    SET v_tanggal_akhir = CONCAT(p_tahun, '-12-31 23:59:59');

    SELECT
        1 AS no,
        'RSUD SOREANG' AS nama_rumah_sakit,

        -- =====================================================
        -- JUMLAH TEMPAT TIDUR
        -- =====================================================

        (
            SELECT COUNT(*)
            FROM beds
            WHERE deleted_at IS NULL AND hidden = 'N'
        ) AS jumlah_tempat_tidur,

        -- =====================================================
        -- PASIEN KELUAR (HIDUP + MATI)
        -- =====================================================

        SUM(
            CASE
                WHEN ri.tgl_keluar IS NOT NULL
                AND p.kelamin = 'L'
                THEN 1 ELSE 0
            END
        ) AS keluar_l,

        SUM(
            CASE
                WHEN ri.tgl_keluar IS NOT NULL
                AND p.kelamin = 'P'
                THEN 1 ELSE 0
            END
        ) AS keluar_p,

        SUM(
            CASE
                WHEN ri.tgl_keluar IS NOT NULL
                THEN 1 ELSE 0
            END
        ) AS keluar_total,

        -- =====================================================
        -- PASIEN KELUAR MATI
        -- =====================================================

        SUM(
            CASE
                WHEN r.kondisi_akhir_pasien IN ('8','9','16')
                AND p.kelamin = 'L'
                THEN 1 ELSE 0
            END
        ) AS mati_l,

        SUM(
            CASE
                WHEN r.kondisi_akhir_pasien IN ('8','9','16')
                AND p.kelamin = 'P'
                THEN 1 ELSE 0
            END
        ) AS mati_p,

        SUM(
            CASE
                WHEN r.kondisi_akhir_pasien IN ('8','9','16')
                THEN 1 ELSE 0
            END
        ) AS mati_total,

        -- =====================================================
        -- MATI >= 48 JAM
        -- =====================================================

        SUM(
            CASE
                WHEN r.kondisi_akhir_pasien IN ('8','9','16')
                AND TIMESTAMPDIFF(HOUR, ri.tgl_masuk, ri.tgl_keluar) >= 48
                AND p.kelamin = 'L'
                THEN 1 ELSE 0
            END
        ) AS mati48_l,

        SUM(
            CASE
                WHEN r.kondisi_akhir_pasien IN ('8','9','16')
                AND TIMESTAMPDIFF(HOUR, ri.tgl_masuk, ri.tgl_keluar) >= 48
                AND p.kelamin = 'P'
                THEN 1 ELSE 0
            END
        ) AS mati48_p,

        SUM(
            CASE
                WHEN r.kondisi_akhir_pasien IN ('8','9','16')
                AND TIMESTAMPDIFF(HOUR, ri.tgl_masuk, ri.tgl_keluar) >= 48
                THEN 1 ELSE 0
            END
        ) AS mati48_total,

        -- =====================================================
        -- GDR
        -- =====================================================

        ROUND(
						(
								SUM(
										CASE
												WHEN r.kondisi_akhir_pasien IN ('8','9','16')
												AND p.kelamin = 'L'
												THEN 1 ELSE 0
										END
								) * 1000
						) /
						NULLIF(
								SUM(
										CASE
												WHEN ri.tgl_keluar IS NOT NULL
												AND p.kelamin = 'L'
												THEN 1 ELSE 0
										END
								),
								0
						),
						2
				) AS gdr_l,
				
				ROUND(
						(
								SUM(
										CASE
												WHEN r.kondisi_akhir_pasien IN ('8','9','16')
												AND p.kelamin = 'P'
												THEN 1 ELSE 0
										END
								) * 1000
						) /
						NULLIF(
								SUM(
										CASE
												WHEN ri.tgl_keluar IS NOT NULL
												AND p.kelamin = 'P'
												THEN 1 ELSE 0
										END
								),
								0
						),
						2
				) AS gdr_p,
				
				ROUND(
						(
								SUM(
										CASE
												WHEN r.kondisi_akhir_pasien IN ('8','9','16')
												THEN 1 ELSE 0
										END
								) * 1000
						) /
						NULLIF(
								SUM(
										CASE
												WHEN ri.tgl_keluar IS NOT NULL
												THEN 1 ELSE 0
										END
								),
								0
						),
						2
				) AS gdr_total,



        -- =====================================================
        -- NDR
        -- =====================================================

        ROUND(
						(
								SUM(
										CASE
												WHEN r.kondisi_akhir_pasien IN ('8','9','16')
												AND TIMESTAMPDIFF(HOUR, ri.tgl_masuk, ri.tgl_keluar) >= 48
												AND p.kelamin = 'L'
												THEN 1 ELSE 0
										END
								) * 1000
						) /
						NULLIF(
								(
										SUM(
												CASE
														WHEN ri.tgl_keluar IS NOT NULL
														AND p.kelamin = 'L'
														THEN 1 ELSE 0
												END
										)
										-
										SUM(
												CASE
														WHEN r.kondisi_akhir_pasien IN ('8','9','16')
														AND TIMESTAMPDIFF(HOUR, ri.tgl_masuk, ri.tgl_keluar) < 48
														AND p.kelamin = 'L'
														THEN 1 ELSE 0
												END
										)
								),
								0
						),
						2
				) AS ndr_l,
				
				ROUND(
						(
								SUM(
										CASE
												WHEN r.kondisi_akhir_pasien IN ('8','9','16')
												AND TIMESTAMPDIFF(HOUR, ri.tgl_masuk, ri.tgl_keluar) >= 48
												AND p.kelamin = 'P'
												THEN 1 ELSE 0
										END
								) * 1000
						) /
						NULLIF(
								(
										SUM(
												CASE
														WHEN ri.tgl_keluar IS NOT NULL
														AND p.kelamin = 'P'
														THEN 1 ELSE 0
												END
										)
										-
										SUM(
												CASE
														WHEN r.kondisi_akhir_pasien IN ('8','9','16')
														AND TIMESTAMPDIFF(HOUR, ri.tgl_masuk, ri.tgl_keluar) < 48
														AND p.kelamin = 'P'
														THEN 1 ELSE 0
												END
										)
								),
								0
						),
						2
				) AS ndr_p,
				
				ROUND(
						(
								SUM(
										CASE
												WHEN r.kondisi_akhir_pasien IN ('8','9','16')
												AND TIMESTAMPDIFF(HOUR, ri.tgl_masuk, ri.tgl_keluar) >= 48
												
												THEN 1 ELSE 0
										END
								) * 1000
						) /
						NULLIF(
								(
										SUM(
												CASE
														WHEN ri.tgl_keluar IS NOT NULL
														
														THEN 1 ELSE 0
												END
										)
										-
										SUM(
												CASE
														WHEN r.kondisi_akhir_pasien IN ('8','9','16')
														AND TIMESTAMPDIFF(HOUR, ri.tgl_masuk, ri.tgl_keluar) < 48
														
														THEN 1 ELSE 0
												END
										)
								),
								0
						),
						2
				) AS ndr_total

    FROM rawatinaps ri
    JOIN registrasis r ON r.id = ri.registrasi_id
    JOIN pasiens p ON p.id = r.pasien_id

    WHERE ri.tgl_keluar >= v_tanggal_awal
      AND ri.tgl_keluar <= v_tanggal_akhir
      AND ri.deleted_at IS NULL;

END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `laporan_rm_profil_56` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `laporan_rm_profil_56`(
    IN p_tahun INT
)
BEGIN

    DECLARE v_tanggal_awal DATETIME;
    DECLARE v_tanggal_akhir DATETIME;
    DECLARE v_jumlah_hari INT;

    SET v_tanggal_awal = CONCAT(p_tahun, '-01-01 00:00:00');
    SET v_tanggal_akhir = CONCAT(p_tahun, '-12-31 23:59:59');

    SET v_jumlah_hari = DAYOFYEAR(CONCAT(p_tahun, '-12-31'));

    SELECT
        1 AS no,
        'RSUD SOREANG' AS nama_rumah_sakit,

        -- =====================================================
        -- JUMLAH TEMPAT TIDUR
        -- =====================================================

        (
            SELECT COUNT(*)
            FROM beds
            WHERE deleted_at IS NULL
						AND hidden = 'N'
        ) AS jumlah_tempat_tidur,

        -- =====================================================
        -- PASIEN KELUAR (HIDUP + MATI)
        -- =====================================================

        COUNT(
            CASE
                WHEN ri.tgl_keluar IS NOT NULL THEN 1
            END
        ) AS pasien_keluar,

        -- =====================================================
        -- JUMLAH HARI PERAWATAN
        -- =====================================================

        SUM(
            DATEDIFF(
                DATE(IFNULL(ri.tgl_keluar, NOW())),
                DATE(ri.tgl_masuk)
            )
        ) AS jumlah_hari_perawatan,

        -- =====================================================
        -- JUMLAH LAMA DIRAWAT
        -- =====================================================

        AVG(
            DATEDIFF(
                DATE(IFNULL(ri.tgl_keluar, NOW())),
                DATE(ri.tgl_masuk)
            )
        ) AS jumlah_lama_dirawat,

        -- =====================================================
        -- BOR
        -- =====================================================

        ROUND(
            (
                SUM(
                    DATEDIFF(
                        DATE(IFNULL(ri.tgl_keluar, NOW())),
                        DATE(ri.tgl_masuk)
                    )
                )
                /
                (
                    (
                        SELECT COUNT(*)
                        FROM beds
                        WHERE deleted_at IS NULL
                    )
                    * v_jumlah_hari
                )
            ) * 100,
            2
        ) AS bor,

        -- =====================================================
        -- BTO
        -- =====================================================

        ROUND(
            (
                COUNT(
                    CASE
                        WHEN ri.tgl_keluar IS NOT NULL THEN 1
                    END
                )
                /
                (
                    SELECT COUNT(*)
                    FROM beds
                    WHERE deleted_at IS NULL
                )
            ),
            2
        ) AS bto,

        -- =====================================================
        -- TOI
        -- =====================================================

        ROUND(
            (
                (
                    (
                        SELECT COUNT(*)
                        FROM beds
                        WHERE deleted_at IS NULL
                    )
                    * v_jumlah_hari
                )
                -
                SUM(
                    DATEDIFF(
                        DATE(IFNULL(ri.tgl_keluar, NOW())),
                        DATE(ri.tgl_masuk)
                    )
                )
            )
            /
            NULLIF(
                COUNT(
                    CASE
                        WHEN ri.tgl_keluar IS NOT NULL THEN 1
                    END
                ),
                0
            ),
            2
        ) AS toi,

        -- =====================================================
        -- ALOS
        -- =====================================================

        ROUND(
            (
                SUM(
                    DATEDIFF(
                        DATE(IFNULL(ri.tgl_keluar, NOW())),
                        DATE(ri.tgl_masuk)
                    )
                )
                /
                NULLIF(
                    COUNT(
                        CASE
                            WHEN ri.tgl_keluar IS NOT NULL THEN 1
                        END
                    ),
                    0
                )
            ),
            2
        ) AS alos

    FROM rawatinaps ri

    WHERE ri.tgl_masuk >= v_tanggal_awal
      AND ri.tgl_masuk <= v_tanggal_akhir
      AND ri.deleted_at IS NULL;

END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-07-17 18:42:22
