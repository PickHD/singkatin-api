const router = require("express").Router();

const { Short } = require("../models");
const short = require("./short.route");
const requester = require("./requester.route");

const { fetchLimiter } = require("../middlewares/rateLimitHandler.middleware");

router.get("/", [fetchLimiter], async (req, res, next) => {
    try {
        const countShortUrl = await Short.count({});
        return res.status(200).json({
            title: "Singkatin API Gateway",
            success: true,
            API_KEY_HEADER_NAME: "X-API-KEY",
            routes: [
                {
                    name: "Short URL API routes",
                    total: 5,
                    refLink: [
                        {
                            method: "GET",
                            url: `${req.protocol}://${req.headers.host}/api/v1/url`,
                            description: "Retrieve all url shortened",
                            status: {
                                code: [200, 401, 429, 500],
                                cached: true,
                                middleware: true,
                                apiKeyRequired: true
                            }

                        },
                        {
                            method: "POST",
                            url: `${req.protocol}://${req.headers.host}/api/v1/url`,
                            description: "Create a url shortened",
                            status: {
                                code: [201, 400, 401, 429, 500],
                                cached: false,
                                middleware: true,
                                apiKeyRequired: true
                            },
                            body: {
                                fields: [{ name: "full_url", type: "string", maxLength: 255 }, { name: "short_url", type: "string", maxLength: 255 }],
                                required: true
                            }

                        },
                        {
                            method: "GET",
                            url: `${req.protocol}://${req.headers.host}/api/v1/url/{shortUrl}`,
                            description: "Redirecting to real url selected by shortUrl in parameter itself",
                            status: {
                                code: [200, 401, 404, 429, 500],
                                cached: true,
                                middleware: true,
                                apiKeyRequired: true
                            },
                            parameter: {
                                name: "shortUrl",
                                type: "string",
                                fixLength: 10,
                                apiKeyRequired: true
                            }
                        },
                        {
                            method: "DELETE",
                            url: `${req.protocol}://${req.headers.host}/api/v1/url/{shortUrl}`,
                            description: "delete 1 short url by shortUrl in parameter itself",
                            status: {
                                code: [200, 401, 404, 500],
                                cached: false,
                                middleware: true,
                                apiKeyRequired: true
                            },
                            parameter: {
                                name: "shortUrl",
                                type: "string",
                                fixLength: 10,
                                required: true
                            }

                        }
                    ]
                },
                {
                    name: "Request API Key Routes",
                    total: 1,
                    refLink: [
                        {
                            method: "POST",
                            url: `${req.protocol}://${req.headers.host}/api/v1/request-api-key`,
                            description: "Requesting API Key represented by a encrypted tokens",
                            status: {
                                code: [201, 400, 429, 500],
                                cached: false,
                                middleware: true,
                                apiKeyRequired: false
                            },
                            body: {
                                fields: [{ name: "email", type: "string", minLength: 6 }],
                                required: true
                            }
                        },

                    ]
                }
            ],
            status: {
                total_shorted_url: countShortUrl
            }

        });
    } catch (e) {
        res.statusCode = 500;
        return next(e);
    }

});

//! V1 API NESTED ROUTES
router.use("/api/v1/url", short);
router.use("/api/v1/request-api-key", requester);

module.exports = router;