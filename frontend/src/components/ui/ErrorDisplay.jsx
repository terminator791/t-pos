import React from "react";
import Card from "@/components/ui/Card";
import Button from "@/components/ui/Button";
import Icon from "@/components/ui/Icon";

const ErrorDisplay = ({ 
  message = "Something went wrong", 
  onRetry, 
  showRetry = true,
  className = "" 
}) => {
  return (
    <Card className={`text-center p-8 ${className}`}>
      <div className="flex flex-col items-center space-y-4">
        <div className="h-12 w-12 rounded-full bg-red-100 dark:bg-red-900/20 flex items-center justify-center">
          <Icon icon="ph:warning" className="text-2xl text-red-600 dark:text-red-400" />
        </div>
        <div>
          <h3 className="text-lg font-medium text-gray-800 dark:text-white">
            Error
          </h3>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            {message}
          </p>
        </div>
        {showRetry && onRetry && (
          <Button
            onClick={onRetry}
            className="btn-primary"
            size="sm"
          >
            <Icon icon="ph:arrow-clockwise" className="mr-2" />
            Try Again
          </Button>
        )}
      </div>
    </Card>
  );
};

export default ErrorDisplay;