import fg from "fast-glob";

export async function scanFiles() {

  let config;
  try {
    const imported = await import("../i18n.config.ts");
    config = imported.default;
  } catch (err) {
    throw new Error("i18n.config.ts not exists or format error");
  }

  const { entryDirs, extensions } = config;
  if (!entryDirs || !extensions) {
    throw new Error("i18n.config.ts missing entryDirs/extensions config");
  }

  const files = await fg(
    entryDirs.map((dir: string) => `${dir}/**/*.{${extensions.join(",")}}`)
  );
  return files;
}