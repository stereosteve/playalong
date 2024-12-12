import Multitrack, { type MultitrackTracks } from "wavesurfer-multitrack";

type Song = {
  id: string;
  title: string;
  genre: string;
  stems: {
    name: string;
    path: string;
  }[];
};

const j = document.getElementById("SongData")?.textContent;
if (!j) {
  throw new Error("missing song data");
}

const song = JSON.parse(j) as Song;
console.log("song", song);

const stems: MultitrackTracks = song.stems.map((s) => {
  return {
    id: s.name,
    url: s.path,
    startPosition: 0,
    intro: {
      endTime: -1,
      label: s.name,
    },
  };
});

const multitrack = Multitrack.create(stems, {
  container: document.querySelector("#container")!,
});

function togglePlay() {
  multitrack.isPlaying() ? multitrack.pause() : multitrack.play();
  playButton.textContent = multitrack.isPlaying() ? "Pause" : "Play";
}

const playButton = document.querySelector("#play") as HTMLButtonElement;
const restartButton = document.querySelector("#restart") as HTMLButtonElement;

multitrack.once("canplay", () => {
  // buttons
  playButton.disabled = false;
  playButton.onclick = togglePlay;

  restartButton.disabled = false;
  restartButton.onclick = () => {
    multitrack.setTime(0);
  };

  // keyboard listener
  document.addEventListener("keyup", (e) => {
    if (e.code === "Space") togglePlay();
  });

  const mixer = document.querySelector("#mixer")!;
  song.stems.forEach((stem, idx) => {
    const mixerRow = document.createElement("div");
    mixerRow.classList.add("mixer-row");

    const label = document.createElement("b");
    label.innerText = stem.name;

    const slider = document.createElement("input");
    slider.type = "range";
    slider.id = "slider";
    slider.min = "0";
    slider.max = "1";
    slider.value = "1";
    slider.step = "0.01";
    slider.oninput = (ev: any) => {
      const val = parseFloat(ev.target.value);
      multitrack.setTrackVolume(idx, val);
    };

    mixerRow.append(label, slider);
    mixer.append(mixerRow);
  });
});
