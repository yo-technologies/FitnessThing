import { Spinner } from "@heroui/react";

export const Loading = ({
  showText = true,
  size,
}: {
  showText?: boolean;
  size?: "sm" | "md" | "lg";
}) => {
  return (
    <div className="p4 flex flex-col flex-grow items-center justify-center gap-4">
      <Spinner size={size} />
      {showText && <p className="text-sm text-neutral-600">Загрузка...</p>}
    </div>
  );
};
