import React from "react";
import Textinput from "@/components/ui/Textinput";
import Textarea from "@/components/ui/Textarea";
import Select from "@/components/ui/Select";

const FormField = ({
  name,
  label,
  type = "text",
  placeholder,
  register,
  error,
  options = [],
  disabled = false,
  required = false,
  ...props
}) => {
  const baseProps = {
    name,
    label,
    placeholder,
    register,
    error,
    disabled,
    ...props,
  };

  switch (type) {
    case "textarea":
      return <Textarea {...baseProps} />;
    
    case "select":
      return (
        <Select
          {...baseProps}
          options={options}
        />
      );
    
    case "number":
      return (
        <Textinput
          {...baseProps}
          type="number"
          step="any"
        />
      );
    
    case "email":
      return (
        <Textinput
          {...baseProps}
          type="email"
        />
      );
    
    case "password":
      return (
        <Textinput
          {...baseProps}
          type="password"
        />
      );
    
    default:
      return (
        <Textinput
          {...baseProps}
          type={type}
        />
      );
  }
};

export default FormField;