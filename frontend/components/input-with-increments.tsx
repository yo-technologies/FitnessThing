import { Button } from "@nextui-org/button";
import { Input } from "@nextui-org/input";
import clsx from "clsx";
import { useEffect, useState } from "react";

function getStepDecimals(step: number) {
  const s = step.toString();
  const idx = s.indexOf(".");

  return idx === -1 ? 0 : s.length - idx - 1;
}

function snapToStep(value: number, step: number, min?: number) {
  if (Number.isNaN(value)) return min ?? 0;
  const snapped = Math.round(value / step) * step;
  const decimals = getStepDecimals(step);
  let result = Number(snapped.toFixed(decimals));

  if (typeof min === "number" && result < min) result = min;

  return result;
}

export function IncrementButtons({
  onPress,
  isSubtract,
  radius,
  className,
}: {
  onPress: () => void;
  isSubtract?: boolean;
  radius?: "sm" | "md" | "lg";
  className?: string;
}) {
  return (
    <div className={clsx(className, "flex flex-col h-full")}>
      <Button
        isIconOnly
        className="min-w-fit h-full w-full max-w-full"
        radius={radius}
        size={radius}
        onPress={onPress}
      >
        {isSubtract ? "-" : "+"}
      </Button>
    </div>
  );
}

function normalize(str: string) {
  return str.replace(",", ".");
}

function isIntermediateInput(normalized: string, allowFloat: boolean) {
  if (allowFloat) {
    return (
      normalized === "" ||
      normalized === "." ||
      normalized === "-" ||
      normalized === "-." ||
      normalized.endsWith(".")
    );
  }

  // int-only: допускаем только пустую строку и минус как промежуточные
  return normalized === "" || normalized === "-";
}

function clampMin(num: number, min?: number) {
  if (typeof min === "number" && num < min) return min;

  return num;
}

function useNumericInput({
  value,
  setValue,
  min,
  step = 2.5,
  allowFloat = true,
}: {
  value: number;
  setValue: (v: number) => void;
  min?: number;
  step?: number;
  allowFloat?: boolean;
}) {
  const [inputStr, setInputStr] = useState<string>(String(value));

  useEffect(() => {
    const n = normalize(inputStr);
    const intermediate = isIntermediateInput(n, allowFloat);

    if (!intermediate) {
      const parsed = Number(n);

      if (!Number.isNaN(parsed) && parsed !== value) {
        setInputStr(String(value));
      }
    }
  }, [value, allowFloat]);

  const onChange = (str: string) => {
    // sanitize for int-only: удаляем запятые и точки
    const sanitized = allowFloat ? str : str.replace(/[.,]/g, "");

    setInputStr(sanitized);

    const n = normalize(sanitized);

    if (isIntermediateInput(n, allowFloat)) return;

    const num = allowFloat ? Number(n) : parseInt(n, 10);

    if (Number.isNaN(num)) return;

    const clamped = clampMin(num, min);

    setValue(clamped);
  };

  const onBlur = () => {
    const n = normalize(inputStr).trim();

    if (isIntermediateInput(n, allowFloat)) {
      const fallback = typeof min === "number" ? min : 0;

      setValue(fallback);
      setInputStr(String(fallback));

      return;
    }

    const num = allowFloat ? Number(n) : parseInt(n, 10);

    if (Number.isNaN(num)) {
      setInputStr(String(value));

      return;
    }

    const clamped = clampMin(num, min);

    setValue(clamped);
    setInputStr(String(clamped));
  };

  const inc = () => {
    let next = clampMin(snapToStep(value + step, step, min), min);

    if (!allowFloat) next = Math.round(next);

    setValue(next);
    setInputStr(String(next));
  };

  const dec = () => {
    let next = clampMin(snapToStep(value - step, step, min), min);

    if (!allowFloat) next = Math.round(next);

    setValue(next);
    setInputStr(String(next));
  };

  return { inputStr, onChange, onBlur, inc, dec };
}

export function InputWithIncrement({
  value,
  setValue,
  label,
  labelNode,
  placeholder,
  className,
  size,
  classNames,
  min,
  step = 2.5,
  allowFloat = true,
}: {
  value: number;
  setValue: (value: number) => void;
  label?: string;
  labelNode?: React.ReactNode;
  placeholder: string;
  className?: string;
  size?: "sm" | "md" | "lg";
  classNames?: { incrementButton: string };
  min?: number;
  step?: number;
  allowFloat?: boolean;
}) {
  const { inputStr, onChange, onBlur, inc, dec } = useNumericInput({
    value,
    setValue,
    min,
    step,
    allowFloat,
  });

  return (
    <>
      {labelNode ? labelNode : <p className="text-md font-light">{label}</p>}
      <div className="flex flex-row gap-2 items-center h-full">
        <IncrementButtons
          isSubtract
          className={clsx("h-full", classNames?.incrementButton)}
          radius={size}
          onPress={dec}
        />
        <Input
          isRequired
          className={clsx("p-0 w-full h-full h-8", className)}
          classNames={{ inputWrapper: "h-full max-h-full min-h-fit" }}
          inputMode={allowFloat ? "decimal" : "numeric"}
          placeholder={placeholder}
          size={size}
          type="text"
          value={inputStr}
          onBlur={onBlur}
          onValueChange={onChange}
        />
        <IncrementButtons
          className={clsx("h-full", classNames?.incrementButton)}
          radius={size}
          onPress={inc}
        />
      </div>
    </>
  );
}
