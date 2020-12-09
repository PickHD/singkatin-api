const CronJob = require("cron").CronJob;
const { Token } = require("../models");

console.log("\t\t===== CRONJOB STARTED ===== \n");
const job = new CronJob("00 00 00 * * *", async () => {
    try {
        const d = new Date();
        await Token.destroy({ truncate: true });
        console.log(`=====ALL TOKEN DELETED at: ${d}=====`);
    } catch (e) {
        console.error(e);
    }

});

module.exports = job;