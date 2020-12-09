const router = require("express").Router();

//!IMPORT CONTROLLER & MIDDLEWARE
const { getAllUrl, createShortUrlHandler, getRedirectUrl, delOneShortUrlHandler } = require("../controllers/short.controller");
const { verifyApiKey } = require("../middlewares/verifyApiKey.middleware");

router.get("/", [verifyApiKey], getAllUrl);
router.post("/", [verifyApiKey], createShortUrlHandler);
router.get("/:shortUrl", [verifyApiKey], getRedirectUrl);
router.delete("/:shortUrl", [verifyApiKey], delOneShortUrlHandler);

module.exports = router;