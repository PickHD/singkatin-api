const { Short } = require("../models");
const { customAlphabet } = require("nanoid");
const { validationResult } = require("express-validator");

//!GENERATE CUSTOM ALPHABET NANOID 
const customNanoId =customAlphabet("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ",10);

exports.getAllUrl = async (req, res, next) => {
    try {
        const { page } = req.query;

        const perPage = 10;
        const getPage = page || 1;
        const resOffset = (perPage * getPage) - perPage;

        const { count, rows } = await Short.findAndCountAll({
            attributes: ["full_url", "short_url", "visited"],
            limit: perPage,
            offset: resOffset,
            order: [["createdAt", "ASC"]],
            raw:true
        });

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
                short_url: customNanoId()
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
                await Short.update({ visited: findShortUrl.visited += 1 }, {
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
                await Short.destroy({
                    where: {
                        short_url: shortUrl
                    }
                });
                return res.status(200).json({
                    success: true,
                    data: {
                        msg: `Short URL ${shortUrl} was successfully deleted.`
                    }
                });
            }
        }
    } catch (e) {
        res.statusCode = 500;
        return next(e);
    }
};