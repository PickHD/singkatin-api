"use strict";
module.exports = {
    up: async (queryInterface, Sequelize) => {
        await queryInterface.createTable("Tokens", {
            id: {
                allowNull: false,
                autoIncrement: true,
                primaryKey: true,
                type: Sequelize.INTEGER
            },
            token: {
                type: Sequelize.STRING,
                allowNull: false
            },
            RequesterId: {
                type: Sequelize.INTEGER,
                allowNull: false
            },
            createdAt: {
                allowNull: false,
                type: Sequelize.DATE
            },
            updatedAt: {
                allowNull: false,
                type: Sequelize.DATE
            }
        });
        await queryInterface.addConstraint("Tokens", {
            fields: ["RequesterId"],
            type: "foreign key",
            name: "requester_fkey_constraint",
            references: {
                table: "Requesters",
                field: "id"
            },
            onDelete: "cascade",
            onUpdate: "cascade"
        });
    },
    down: async (queryInterface, Sequelize) => {
        await queryInterface.dropTable("Tokens");
    }
};