const { validationResult } = require("express-validator");

const { Requester, Token } = require("../models");

const { generateToken } = require("../helpers/queryFunction.helper");

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
            //!if email existed in database , remove old token & create token again 
            if (checkExistingEmail) {
                await Token.destroy({ where: { RequesterId: checkExistingEmail.id } });

                const genTokenExistingUser = await generateToken(checkExistingEmail);
                return res.status(201).json({
                    success: true,
                    data: {
                        API_KEY_HEADER_NAME: "X-API-KEY",
                        value: genTokenExistingUser
                    }
                });
            } else {
                //! if not,create new token  
                const createRequester = await Requester.create({
                    email: email
                });

                const genNewToken = await generateToken(createRequester);
                return res.status(201).json({
                    success: true,
                    data: {
                        API_KEY_HEADER_NAME: "X-API-KEY",
                        value: genNewToken
                    }
                });
            }
        }

    } catch (e) {
        res.statusCode = 500;
        return next(e);
    }

};