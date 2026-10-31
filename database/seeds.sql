/*M!999999\- enable the sandbox mode */ 
-- MariaDB dump 10.19  Distrib 10.11.19-MariaDB, for debian-linux-gnu (x86_64)
--
-- Host: localhost    Database: smsagribd_sc
-- ------------------------------------------------------
-- Server version	10.11.19-MariaDB-0+deb12u1

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Dumping data for table `countries`
--

LOCK TABLES `countries` WRITE;
/*!40000 ALTER TABLE `countries` DISABLE KEYS */;
INSERT INTO `countries` VALUES
(1,'Bangladesh','880','BD',1,'2026-10-31 22:31:41','2026-10-31 22:31:41');
/*!40000 ALTER TABLE `countries` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Dumping data for table `operators`
--

LOCK TABLES `operators` WRITE;
/*!40000 ALTER TABLE `operators` DISABLE KEYS */;
INSERT INTO `operators` VALUES
(1,1,'Grameenphone','[\"17\",\"13\"]',0.4500,0.3000,1,'2026-10-31 22:31:41','2026-10-31 22:31:41'),
(2,1,'Banglalink','[\"19\",\"14\"]',0.4500,0.3000,1,'2026-10-31 22:31:41','2026-10-31 22:31:41'),
(3,1,'Robi','[\"18\"]',0.4500,0.3000,1,'2026-10-31 22:31:41','2026-10-31 22:31:41'),
(4,1,'Airtel','[\"16\"]',0.4500,0.3000,1,'2026-10-31 22:31:41','2026-10-31 22:31:41'),
(5,1,'Teletalk','[\"15\"]',0.4500,0.3000,1,'2026-10-31 22:31:41','2026-10-31 22:31:41');
/*!40000 ALTER TABLE `operators` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Dumping data for table `tariff_plans`
--

LOCK TABLES `tariff_plans` WRITE;
/*!40000 ALTER TABLE `tariff_plans` DISABLE KEYS */;
INSERT INTO `tariff_plans` VALUES
(1,'Standard Plan','standard','standard','Default standard rate plan','BDT','active',NULL,'2026-10-31 22:31:41','2026-10-31 22:31:41');
/*!40000 ALTER TABLE `tariff_plans` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Dumping data for table `tariff_rates`
--

LOCK TABLES `tariff_rates` WRITE;
/*!40000 ALTER TABLE `tariff_rates` DISABLE KEYS */;
INSERT INTO `tariff_rates` VALUES
(1,1,1,0.4500,0.3000,0.2500,NULL,'2026-10-31 18:31:41','2026-10-31 22:31:41','2026-10-31 22:31:41'),
(2,1,2,0.4500,0.3000,0.2500,NULL,'2026-10-31 18:31:41','2026-10-31 22:31:41','2026-10-31 22:31:41'),
(3,1,3,0.4500,0.3000,0.2500,NULL,'2026-10-31 18:31:41','2026-10-31 22:31:41','2026-10-31 22:31:41'),
(4,1,4,0.4500,0.3000,0.2500,NULL,'2026-10-31 18:31:41','2026-10-31 22:31:41','2026-10-31 22:31:41'),
(5,1,5,0.4500,0.3000,0.2500,NULL,'2026-10-31 18:31:41','2026-10-31 22:31:41','2026-10-31 22:31:41');
/*!40000 ALTER TABLE `tariff_rates` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Dumping data for table `settings`
--

LOCK TABLES `settings` WRITE;
/*!40000 ALTER TABLE `settings` DISABLE KEYS */;
INSERT INTO `settings` VALUES
(1,'general','appName','\"SMS Aggregator Panel\"',0,'2026-10-31 22:41:04','2026-10-31 22:41:04'),
(2,'general','logoUrl','\"/logo.png?v=aggregator_v2\"',0,'2026-10-31 22:41:04','2026-10-31 22:47:42');
/*!40000 ALTER TABLE `settings` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-10-31 18:50:46
