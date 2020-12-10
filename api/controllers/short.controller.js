const { Short } = require("../models");
const { validationResult } = require("express-validator");

exports.getAllUrl = async (req, res, next) => {
    try {

        //TODO : Create Dynamic Query String For Filter & Searching Rows
        //! const {} = req.query 

        const { count, rows } = await Short.findAndCountAll({ raw: true });

        return res.status(200).json({
            success: true,
            data: {
                count: count,
                rows: rows
            }
        });

    } catch (e) {
        res.statusCode = 500;
        return next(e);
    }

};

exports.createShortUrlHandler = async (req, res, next) => {
    try {
        const { full_url } = req.body;
        const badReqError = validationResult(req);
        if (!badReqError.isEmpty()) {
            return res.status(400).json({ validation_errors: badReqError.array() });
        } else {
            const createShortUrl = await Short.create({
                full_url: full_url,
            });
            return res.status(201).json({
                success: true,
                data: {
                    msg: "URL Shorted Successfully",
                    shortUrlResult: createShortUrl.short_url
                }
            });
        }

    } catch (e) {
        res.statusCode = 500;
        return next(e);
    }
};

exports.getRedirectUrl = async (req, res, next) => {
    try {
        const { shortUrl } = req.params;

        const badReqError = validationResult(req);
        if (!badReqError.isEmpty()) {
            return res.status(400).json({ validation_errors: badReqError.array() });
        } else {

            const findShortUrl = await Short.findOne({
                where: {
                    short_url: shortUrl
                }
            });
            if (!findShortUrl) {
                res.statusCode = 404;
                return next(new Error(`${shortUrl} not found, make sure to check the data first`));
            } else {
                await Short.update({ visited: findShortUrl.visited +=1 }, {
                    where: {
                        short_url: shortUrl
                    }
                });
                return res.status(200).json({
                    success: true,
                    data: {
                        fullUrlResult: findShortUrl.full_url
                    }
                });
            }
        }
    } catch (e) {
        res.statusCode = 500;
        return next(e);
    }

};

exports.delOneShortUrlHandler = async (req, res, next) => {

};