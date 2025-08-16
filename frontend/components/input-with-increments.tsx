import { Button } from "@nextui-org/button";
import { Input } from "@nextui-org/input";
import clsx from "clsx";

export function IncrementButtons({
  value,
  setValue,
  isSubtract,
  radius,
  className,
}: {
  value: number;
  setValue: (value: number) => void;
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
        onPress={() => {
          if (value > 0 && isSubtract) {
            setValue(value - 1);

            return;
          }
          if (!isSubtract) {
            setValue(value + 1);
          }
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
  type,
  className,
  size,
  classNames,
  min,
}: {
  value: number;
  setValue: (value: number) => void;
  label?: string;
  labelNode?: React.ReactNode;
  placeholder: string;
  type: string;
  className?: string;
  size?: "sm" | "md" | "lg";
  classNames?: { incrementButton: string };
  min?: number;
}) {
  return (
    <>
      {labelNode ? labelNode : <p className="text-md font-light">{label}</p>}
      <div className="flex flex-row gap-2 items-center h-full">
        <IncrementButtons
          isSubtract
          className={clsx("h-full", classNames?.incrementButton)}
          radius={size}
          setValue={setValue}
          value={value}
        />
        <Input
          isRequired
          className={clsx("p-0 w-full h-full h-8", className)}
          classNames={{ inputWrapper: "h-full max-h-full min-h-fit" }}
          min={min}
          placeholder={placeholder}
          size={size}
          type={type}
          value={String(value)}
          onValueChange={(value) => setValue(Number(value))}
        />
        <IncrementButtons
          className={clsx("h-full", classNames?.incrementButton)}
          radius={size}
          setValue={setValue}
          value={value}
        />
      </div>
    </>
  );
}
