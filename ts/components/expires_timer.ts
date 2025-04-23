import { Component } from "ts/types";

export default {
  id: "expires_timer",
  mount: (el) => {
    const dateStr = el.getAttribute("data-date") as string;
    const date = new Date(dateStr);

    const formatNumber = (num: number): string => {
      if (num < 10) {
        return `0${num}`;
      }
      return `${num}`;
    };

    const setText = () => {
      let deltaS = (date.getTime() - Date.now()) / 1000;

      const minutes = deltaS / 60;
      deltaS %= 60;

      const seconds = deltaS;

      if (seconds < 0) {
        el.textContent = "00:00";
        return;
      }

      el.textContent = `${formatNumber(Math.floor(minutes))}:${formatNumber(Math.floor(seconds))}`;
    };
    setText();

    const timeout = setInterval(() => {
      setText();
    }, 1e3);

    return () => {
      clearInterval(timeout);
    };
  },
} as Component<HTMLSpanElement>;
