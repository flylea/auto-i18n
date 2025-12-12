import fg from "fast-glob";
import config from "../i18n.config.js";

const { entryDirs, extensions } = config;

export async function scanFiles() {
  const files = await fg(
    entryDirs.map((dir: string) => `${dir}/**/*.{${extensions.join(",")}}`)
  );
  return files;
}
