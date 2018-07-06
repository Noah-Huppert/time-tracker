"use strict";

const path = require("path");
const webpack = require("webpack");

module.exports = {
	entry: {
		"js/main.js": "./src/js/main.js",
		"index.html": "./src/index.html"
	},
	output: {
		path: path.resolve(__dirname, "dist"),
		filename: "[name]"
	},
	module: {
		rules: [
			{
				test: /\.js$/,
				use: [
					{
						loader: "babel-loader",
						options: {
							presets: ["babel-preset-env"]
						}
					}
				]
			},
			{
				test: /\.html$/,
				use: [
					{
						loader: "file-loader",
						options: {
							name: "[name].[ext]"
						}
					}
				]
			}
		]
	},
	stats: {
		colors: true
	},
	devtool: "source-map"
};
