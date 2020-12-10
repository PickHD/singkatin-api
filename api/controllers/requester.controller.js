const crypt = require("crypto");
const { validationResult } = require("express-validator");

const { Token, Requester } = require("../models");

exports.createRequestApiKeyHandler = async (req, res, next) => {
    try {
        const { email } = req.body;

        const badReqError = validationResult(req);

        if (!badReqError.isEmpty()) {
            return res.status(400).json({ validation_errors: badReqError.array() });
        } else {
            const checkExistingEmail = await Requester.findOne({
                where: {
                    email: email
                }
            });
            if (checkExistingEmail) {
                res.statusCode = 400;
                return next(new Error("Email already in used, please to change your email."));
            } else {

                const createRequester = await Requester.create({
                    email: email
                });

                const genToken = await Token.create({
                    token: crypt.randomBytes(23).toString("hex").toUpperCase(),
                    RequesterId: createRequester.id
                });

                return res.status(201).json({
                    success: true,
                    data: {
                        API_KEY_HEADER_NAME: "X-API-KEY",
                        value: genToken.token
                    }
                });
            }
        }

    } catch (e) {
        res.statusCode = 500;
        return next(e);
    }

};