const { VueLoaderPlugin } = require("vue-loader");

const path = require("path");
const webpack = require("webpack");

module.exports = {
  mode: "development",
  devtool: "inline-source-map",
  entry: {
    application: [
      "./app/javascripts/application.js",
      "./app/assets/stylesheets/application.scss",
    ],
  },
  output: {
    filename: "[name].js",
    sourceMapFilename: "[name].js.map",
    path: path.resolve(__dirname, "../../", "./app/assets/builds"),
  },
  module: {
    rules: [
      {
        test: /\.s[ac]ss$/i,
        use: [
          // Creates `style` nodes from JS strings
          { loader: "style-loader", options: { injectType: "linkTag" } },
          // Translates CSS into CommonJS
          { loader: "css-loader"},
          // Compiles Sass to CSS
          { loader: "sass-loader"},
        ],
      },
      {
        test: /\.css$/,
        use: [
          "vue-style-loader",
          {
            loader: "css-loader",
            options: { importLoaders: 1, injectType: "linkTag" },
          },
          "postcss-loader",
        ],
      },
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: {
          loader: "babel-loader"
        },
      },
      {
        test: /\.vue$/,
        loader: "vue-loader",
      },
    ],
  },
  resolve: {
    // alias: {
    //   vue$: "vue/dist/vue.runtime.esm.js",
    // },
    extensions: ["*", ".js", ".vue", ".json", "*.map"],
  },
  optimization: {
    chunkIds: "named",
    minimize: false,

    // concatenateModules: false,
  },

  plugins: [

    new VueLoaderPlugin(),
    new webpack.optimize.LimitChunkCountPlugin({
      maxChunks: 1,
    }),
  ],
};
