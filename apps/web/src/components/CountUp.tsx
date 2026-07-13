import { useInView, useMotionValue, useSpring } from "motion/react";
import { useCallback, useEffect, useRef } from "react";

type CountUpProps = {
  to: number;
  from?: number;
  direction?: "up" | "down";
  delay?: number;
  duration?: number;
  className?: string;
  startWhen?: boolean;
  separator?: string;
  onStart?: () => void;
  onEnd?: () => void;
};

export function CountUp({
  to,
  from = 0,
  direction = "up",
  delay = 0,
  duration = 1.2,
  className = "",
  startWhen = true,
  separator = "",
  onStart,
  onEnd,
}: CountUpProps) {
  const ref = useRef<HTMLSpanElement>(null);
  const motionValue = useMotionValue(direction === "down" ? to : from);
  const springValue = useSpring(motionValue, {
    damping: 20 + 40 * (1 / duration),
    stiffness: 100 * (1 / duration),
  });
  const isInView = useInView(ref, { once: true, margin: "0px" });

  const decimalPlaces = (value: number) => {
    const decimals = value.toString().split(".")[1];
    return decimals && Number.parseInt(decimals, 10) !== 0 ? decimals.length : 0;
  };
  const maxDecimals = Math.max(decimalPlaces(from), decimalPlaces(to));

  const formatValue = useCallback(
    (value: number) => {
      const formatted = new Intl.NumberFormat("en-US", {
        maximumFractionDigits: maxDecimals,
        minimumFractionDigits: maxDecimals,
        useGrouping: Boolean(separator),
      }).format(value);

      return separator ? formatted.replace(/,/g, separator) : formatted;
    },
    [maxDecimals, separator],
  );

  useEffect(() => {
    if (ref.current) ref.current.textContent = formatValue(direction === "down" ? to : from);
  }, [direction, formatValue, from, to]);

  useEffect(() => {
    if (!isInView || !startWhen) return;

    onStart?.();
    const startTimeout = window.setTimeout(() => {
      motionValue.set(direction === "down" ? from : to);
    }, delay * 1000);
    const endTimeout = window.setTimeout(() => onEnd?.(), (delay + duration) * 1000);

    return () => {
      window.clearTimeout(startTimeout);
      window.clearTimeout(endTimeout);
    };
  }, [delay, direction, duration, from, isInView, motionValue, onEnd, onStart, startWhen, to]);

  useEffect(() => springValue.on("change", (value) => {
    if (ref.current) ref.current.textContent = formatValue(value);
  }), [formatValue, springValue]);

  return <span className={className} ref={ref} />;
}
