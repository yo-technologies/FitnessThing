import { Button } from "@nextui-org/button";
import { Input } from "@nextui-org/input";
import clsx from "clsx";

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
  value,
  setValue,
  isSubtract,
  radius,
  className,
  step = 2.5,
  min,
}: {
  value: number;
  setValue: (value: number) => void;
  isSubtract?: boolean;
  radius?: "sm" | "md" | "lg";
  className?: string;
  step?: number;
  min?: number;
}) {
  return (
    <div className={clsx(className, "flex flex-col h-full")}>
      <Button
        isIconOnly
        className="min-w-fit h-full w-full max-w-full"
        radius={radius}
        size={radius}
        onPress={() => {
          if (isSubtract) {
            const next = snapToStep(value - step, step, min);

            if (typeof min === "number" && value <= min) return;
            setValue(next);

            return;
          }
          const next = snapToStep(value + step, step, min);

          setValue(next);
        }}
      >
        {isSubtract ? "-" : "+"}
      </Button>
    </div>
  );
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
}) {
  return (
    <>
      {labelNode ? labelNode : <p className="text-md font-light">{label}</p>}
      <div className="flex flex-row gap-2 items-center h-full">
        <IncrementButtons
          isSubtract
          className={clsx("h-full", classNames?.incrementButton)}
          min={min}
          radius={size}
          setValue={setValue}
          step={step}
          value={value}
        />
        <Input
          isRequired
          className={clsx("p-0 w-full h-full h-8", className)}
          classNames={{ inputWrapper: "h-full max-h-full min-h-fit" }}
          inputMode="decimal"
          placeholder={placeholder}
          size={size}
          type="text"
          value={String(value)}
          onValueChange={(str) => {
            const normalized = str.replace(",", ".");
            const num = Number(normalized);

            if (Number.isNaN(num)) return;
            if (typeof min === "number" && num < min) {
              setValue(min);

              return;
            }
            setValue(num);
          }}
        />
        <IncrementButtons
          className={clsx("h-full", classNames?.incrementButton)}
          min={min}
          radius={size}
          setValue={setValue}
          step={step}
          value={value}
        />
      </div>
    </>
  );
}
