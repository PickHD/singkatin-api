const crypt = require("crypto");
const { Token } = require("../models");

const generateToken = async (requester) => {
    const dummyToken = await Token.create({
        token: crypt.randomBytes(23).toString("hex").toUpperCase(),
        RequesterId: requester.id
    });
    return dummyToken.token;

};

module.exports = {
    generateToken
};