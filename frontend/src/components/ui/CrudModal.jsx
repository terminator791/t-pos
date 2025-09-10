import React from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogClose,
} from "@/components/ui/dialog";
import Button from "@/components/ui/Button";
import Icon from "@/components/ui/Icon";

const CrudModal = ({
  isOpen,
  onClose,
  title,
  description,
  children,
  isLoading = false,
  onSubmit,
  submitText = "Save",
  submitButtonClass = "btn-primary",
  size = "md",
}) => {
  const handleSubmit = (e) => {
    e.preventDefault();
    if (onSubmit) {
      onSubmit(e);
    }
  };

  const sizeClasses = {
    sm: "max-w-md",
    md: "max-w-lg",
    lg: "max-w-xl",
    xl: "max-w-2xl",
    "2xl": "max-w-4xl",
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className={`${sizeClasses[size]} max-h-[90vh] overflow-auto`}>
        <DialogHeader>
          <DialogTitle className="text-lg font-semibold text-gray-800 dark:text-white">
            {title}
          </DialogTitle>
          {description && (
            <DialogDescription className="text-sm text-gray-500 dark:text-gray-400">
              {description}
            </DialogDescription>
          )}
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="py-4">
            {children}
          </div>

          <div className="flex justify-end space-x-2 pt-4 border-t border-gray-200 dark:border-gray-700">
            <DialogClose asChild>
              <Button
                type="button"
                variant="outline"
                disabled={isLoading}
                className="btn-secondary"
              >
                Cancel
              </Button>
            </DialogClose>
            <Button
              type="submit"
              isLoading={isLoading}
              className={submitButtonClass}
            >
              <Icon icon="ph:check" className="mr-2" />
              {submitText}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
};

export default CrudModal;