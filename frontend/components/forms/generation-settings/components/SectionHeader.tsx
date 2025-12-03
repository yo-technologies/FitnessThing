import React from "react";
import { Popover, PopoverContent, PopoverTrigger } from "@heroui/popover";

import { SectionHeaderProps } from "../types";

import { CircleQuestionIcon } from "@/config/icons";

export function SectionHeader({ title, tooltip, badge }: SectionHeaderProps) {
  return (
    <div className="flex flex-row gap-2 items-center justify-between">
      <div className="flex items-center gap-2">
        <p className="text-medium font-medium">{title}</p>
        {badge}
      </div>
      <Popover backdrop="opaque" size="sm">
        <PopoverTrigger>
          <CircleQuestionIcon className="w-4 h-4 text-default-500 cursor-pointer" />
        </PopoverTrigger>
        <PopoverContent>
          <p className="text-xs font-light text-default-600 p-1">{tooltip}</p>
        </PopoverContent>
      </Popover>
    </div>
  );
}
