// Require
const path = require("path");
const webpack = require("webpack");

const CleanWebpackPlugin = require("clean-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");

// Constants
const buildDir = path.resolve(__dirname, "dist")
const devHttpPort = 5000;

module.exports = {
	// Input
	entry: ["./src/ts/app.tsx"],

	// Output
	output: {
		path: buildDir,
		filename: '[name].[hash].js',
	},	

	// Pre-Processors
	module: {
		rules: [
			// Typescript
			{
				test: /\.tsx$/,
				exclude: [/node_modules/],
				use: [
					{
						loader: "ts-loader"
					}
				]
			},

			// HTML
			{
				test: /\.html$/,
				use: [
					{
						loader: "html-loader"
					}
				]
			},

			// CSSk
			{
				test: /\.css$/,
				use: [
					{
						loader: "css-loader"
					}
				]
			}
		]
	},

	// Plugins
	plugins: [
		// Clean build directory
		new CleanWebpackPlugin([buildDir]),

		// HTML generation
		new HtmlWebpackPlugin({
			title: "Time Tracker",
			meta: {
				viewport: "width=device-with, initial-scale=1"
			},
			template: "src/index.html"
		})
	],

	// Run configuration
	stats: {
		colors: true
	},

	// Development server
	devtool: "source-map",
};
