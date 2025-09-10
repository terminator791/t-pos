import React from "react";
import Icon from "@/components/ui/Icon";

const LoadingSpinner = ({ 
  size = "md", 
  message = "Loading...", 
  className = "",
  showMessage = true 
}) => {
  const sizeClasses = {
    sm: "text-sm",
    md: "text-lg",
    lg: "text-2xl",
    xl: "text-4xl",
  };

  return (
    <div className={`flex flex-col items-center justify-center p-8 ${className}`}>
      <Icon 
        icon="ph:spinner" 
        className={`animate-spin text-indigo-500 ${sizeClasses[size]} mb-2`} 
      />
      {showMessage && (
        <p className="text-sm text-gray-500 dark:text-gray-400">
          {message}
        </p>
      )}
    </div>
  );
};

export default LoadingSpinner;