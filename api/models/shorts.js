"use strict";
const {
    Model
} = require("sequelize");
module.exports = (sequelize, DataTypes) => {
    class Shorts extends Model {
    /**
     * Helper method for defining associations.
     * This method is not a part of Sequelize lifecycle.
     * The `models/index` file will call this method automatically.
     */
        static associate(models) {
            // define association here
        }
    }
    Shorts.init({
        full_url: DataTypes.STRING,
        short_url:DataTypes.STRING,
        clicked:DataTypes.INTEGER
    }, {
        sequelize,
        modelName: "Shorts",
    });
    return Shorts;
};