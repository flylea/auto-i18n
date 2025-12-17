import cliProgress from "cli-progress";
import colors from "colors";

let bar: cliProgress.SingleBar | null = null;

export function progressStart(total: number) {
  if (bar) return;

  bar = new cliProgress.SingleBar(
    {
      format:
        colors.cyan("Progress") +
        " |" +
        colors.green("{bar}") +
        "| " +
        colors.yellow("{percentage}%"),
      barCompleteChar: "█",
      barIncompleteChar: "░",
      hideCursor: true,
      barsize: 30,
    },
    cliProgress.Presets.shades_classic
  );

  bar.start(total, 0);
}

export function progressInc(step = 1) {
  if (!bar) return;
  bar.increment(step);
}

export function progressStop() {
  if (!bar) return;
  bar.stop();
  bar = null;
}
