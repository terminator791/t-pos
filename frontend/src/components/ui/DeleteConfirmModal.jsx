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

const DeleteConfirmModal = ({
  isOpen,
  onClose,
  onConfirm,
  title = "Confirm Delete",
  message = "Are you sure you want to delete this item? This action cannot be undone.",
  itemName = "",
  isLoading = false,
}) => {
  const handleConfirm = () => {
    if (onConfirm) {
      onConfirm();
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <div className="flex items-center space-x-3">
            <div className="h-12 w-12 rounded-full bg-red-100 dark:bg-red-900/20 flex items-center justify-center">
              <Icon icon="ph:warning" className="text-2xl text-red-600 dark:text-red-400" />
            </div>
            <div>
              <DialogTitle className="text-lg font-semibold text-gray-800 dark:text-white">
                {title}
              </DialogTitle>
            </div>
          </div>
        </DialogHeader>

        <div className="py-4">
          <DialogDescription className="text-sm text-gray-600 dark:text-gray-400">
            {message}
          </DialogDescription>
          {itemName && (
            <div className="mt-3 p-3 bg-gray-50 dark:bg-gray-800 rounded-md">
              <p className="text-sm font-medium text-gray-800 dark:text-white">
                Item: <span className="font-normal">{itemName}</span>
              </p>
            </div>
          )}
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
            type="button"
            onClick={handleConfirm}
            isLoading={isLoading}
            className="btn-danger"
          >
            <Icon icon="ph:trash" className="mr-2" />
            Delete
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default DeleteConfirmModal;