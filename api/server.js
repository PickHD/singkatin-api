const express = require("express");
const cors = require("cors");
const helmet = require("helmet");
const morgan = require("morgan");
const apiCache = require("apicache");
require("dotenv").config();

const app = express();
const port = process.env.PORT || 8080;
const db = require("./models/index");

const api = require("./routes/index.route");

(async function setupAPI() {
    try {
        let cache = apiCache.options({
            statusCodes: {
                exclude: [400, 404, 403],
                include: [200],
            }
        }).middleware;
        
        app.use(express.json());
        app.use(express.urlencoded({ extended: true }));

        app.use(cors());
        app.use(helmet());

        if (process.env.NODE_ENV === "production") {
            app.use(morgan("short"));
            app.use(cache("30 min"));
            await db.sequelize.authenticate();
        } else {
            app.use(morgan("dev"));
            db.sequelize.sync({});
        }

        app.use("/", api);
    } catch (e) {
        console.error(e);
    }

}());

app.listen(port, () => console.log(`Server is Listening on Port :${port}`));