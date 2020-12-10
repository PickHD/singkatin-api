const router = require("express").Router();
const { body } = require("express-validator");

//!IMPORT CONTROLLER & MIDDLEWARE
const { getAllUrl, createShortUrlHandler, getRedirectUrl, delOneShortUrlHandler } = require("../controllers/short.controller");
const verifyApiKey = require("../middlewares/verifyApiKey.middleware");
const { fetchLimiter, createLimiter } = require("../middlewares/rateLimitHandler.middleware");



router.get("/", [verifyApiKey, fetchLimiter], getAllUrl);

router.post("/", [verifyApiKey, createLimiter,body("full_url").not().isEmpty().isLength({max:255}).isURL()], createShortUrlHandler);

// router.get("/:shortUrl", [verifyApiKey,createLimiter], getRedirectUrl);
// router.delete("/:shortUrl", [verifyApiKey], delOneShortUrlHandler);

module.exports = router;