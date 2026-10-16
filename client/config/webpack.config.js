const path = require("path");
const fs = require("fs");
const webpack = require("webpack");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const TerserPlugin = require("terser-webpack-plugin");
const { CleanWebpackPlugin } = require("clean-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const CssMinimizerPlugin = require("css-minimizer-webpack-plugin");
const LiveReloadPlugin = require("webpack-livereload-plugin");
const CopyWebpackPlugin = require("copy-webpack-plugin");

const rootDir = path.resolve(__dirname, "../../");

const appEnv = process.env.APP_ENV || process.env.NODE_ENV || "dev";

const envFileMap = {
  dev: ".env.dev",
  development: ".env.dev",
  stag: ".env.stag",
  staging: ".env.stag",
  prod: ".env.prod",
  production: ".env.prod",
};

const selectedEnvFile = envFileMap[appEnv] || `.env.${appEnv}`;
const envPath = path.resolve(rootDir, selectedEnvFile);

if (fs.existsSync(envPath)) {
  require("dotenv").config({ path: envPath });
} else {
  require("dotenv").config({ path: path.resolve(rootDir, ".env") });
}

const isDev = appEnv === "dev" || appEnv === "development";
const mode = isDev ? "development" : "production";

const plugins = [
  new CleanWebpackPlugin({
    cleanOnceBeforeBuildPatterns: [path.join(rootDir, "static")],
  }),
  new CopyWebpackPlugin({
    patterns: [
      {
        from: path.join(rootDir, "client/public"),
        to: path.join(rootDir, "static"),
        globOptions: {
          ignore: ["**/index.html"],
        },
        noErrorOnMissing: true,
      },
    ],
  }),
  new MiniCssExtractPlugin({
    filename: "css/[name].[contenthash:8].css",
  }),
  new HtmlWebpackPlugin({
    inject: true,
    template: path.join(rootDir, "client/public/index.html"),
    minify: !isDev
      ? {
          html5: true,
          collapseWhitespace: true,
          minifyCSS: true,
          minifyJS: true,
          minifyURLs: false,
          removeComments: true,
        }
      : false,
  }),
  new LiveReloadPlugin({
    appendScriptTag: isDev,
    port: 0,
  }),
  new webpack.DefinePlugin({
    "process.env.APPLICATION_SERVER": JSON.stringify(
      process.env.APPLICATION_SERVER || "",
    ),
    "process.env.APP_ENV": JSON.stringify(process.env.APP_ENV || appEnv),
    "process.env.KEYCLOAK_URL": JSON.stringify(
      process.env.REACT_APP_KEYCLOAK_URL || process.env.KEYCLOAK_URL || "",
    ),
    "process.env.KEYCLOAK_REALM": JSON.stringify(
      process.env.REACT_APP_KEYCLOAK_REALM || process.env.KEYCLOAK_REALM || "",
    ),
    "process.env.KEYCLOAK_CLIENT_ID": JSON.stringify(
      process.env.REACT_APP_KEYCLOAK_CLIENT_ID ||
        process.env.KEYCLOAK_CLIENT_ID ||
        "",
    ),
  }),
];

module.exports = {
  mode,
  devtool: isDev ? "eval-source-map" : "source-map",
  entry: [path.join(rootDir, "client/src/index.tsx")],
  output: {
    path: path.join(rootDir, "static"),
    publicPath: "/",
    filename: isDev ? "js/[name].js" : "js/[name].[contenthash:8].js",
  },
  resolve: {
    extensions: [".tsx", ".ts", ".jsx", ".js", ".json"],
  },
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        exclude: /node_modules/,
        use: {
          loader: "ts-loader",
          options: {
            transpileOnly: true,
          },
        },
      },
      {
        test: /\.(sa|sc|c)ss$/,
        use: [
          isDev ? "style-loader" : MiniCssExtractPlugin.loader,
          "css-loader",
          "sass-loader",
        ],
      },
      {
        test: /\.(png|jpe?g|gif|svg|webp)$/i,
        type: "asset/resource",
        generator: {
          filename: "images/[name][ext]",
        },
      },
      {
        test: /\.(woff2?|eot|ttf|otf)$/i,
        type: "asset/resource",
        generator: {
          filename: "fonts/[name][ext]",
        },
      },
    ],
  },
  optimization: {
    minimize: !isDev,
    minimizer: [
      new TerserPlugin({
        terserOptions: {
          format: {
            comments: false,
          },
        },
        extractComments: false,
      }),
      new CssMinimizerPlugin(),
    ],
  },
  plugins,
};
