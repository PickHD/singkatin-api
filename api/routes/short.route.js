const router = require("express").Router();

//!IMPORT CONTROLLER & MIDDLEWARE
const { getAllUrl, createShortUrlHandler, getRedirectUrl, delOneShortUrlHandler } = require("../controllers/short.controller");
const verifyApiKey = require("../middlewares/verifyApiKey.middleware");
const { fetchLimiter } = require("../middlewares/rateLimitHandler.middleware");

router.get("/", [verifyApiKey, fetchLimiter], getAllUrl);
// router.post("/", [verifyApiKey], createShortUrlHandler);
// router.get("/:shortUrl", [verifyApiKey,createLimiter], getRedirectUrl);
// router.delete("/:shortUrl", [verifyApiKey], delOneShortUrlHandler);

module.exports = router;