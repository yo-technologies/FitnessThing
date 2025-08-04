import React from "react";
import { Card, CardBody } from "@nextui-org/card";
import { Textarea } from "@nextui-org/input";

import { TextFieldProps } from "../types";

import { SectionHeader } from "./SectionHeader";

export function TextField({
  title,
  tooltip,
  value,
  placeholder,
  minRows = 2,
  maxRows = 4,
  badge,
  onChange,
  onBlur,
}: TextFieldProps) {
  return (
    <Card>
      <CardBody className="flex flex-col gap-4">
        <SectionHeader badge={badge} title={title} tooltip={tooltip} />
        <Textarea
          classNames={{
            input:
              "text-sm placeholder:text-default-400 placeholder:text-xs placeholder:font-light",
          }}
          maxRows={maxRows}
          minRows={minRows}
          placeholder={placeholder}
          value={value}
          variant="bordered"
          onBlur={onBlur}
          onValueChange={onChange}
        />
      </CardBody>
    </Card>
  );
}
