import React from "react";
import Icon from "@/components/ui/Icon";
import Button from "@/components/ui/Button";

const CartItem = ({ item, onQuantityChange, onRemove }) => {
  const handleDecrease = () => {
    onQuantityChange(item, item.quantity - 1);
  };

  const handleIncrease = () => {
    onQuantityChange(item, item.quantity + 1);
  };

  const handleRemove = () => {
    onRemove(item.id);
  };

  const itemTotal = (item.quantity * item.unit_price).toFixed(2);

  return (
    <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
      <div className="flex-1 min-w-0">
        <h4 className="font-medium text-sm text-gray-900 dark:text-white truncate">
          {item.product_name || `Product ${item.product_id}`}
        </h4>
        <div className="flex items-center justify-between mt-1">
          <p className="text-xs text-gray-500 dark:text-gray-400">
            ${item.unit_price?.toFixed(2)} each
          </p>
          <p className="text-sm font-medium text-gray-900 dark:text-white">
            ${itemTotal}
          </p>
        </div>
      </div>
      
      <div className="flex items-center space-x-2 ml-4">
        <div className="flex items-center space-x-1 bg-white dark:bg-gray-700 rounded-lg border border-gray-200 dark:border-gray-600">
          <Button
            size="sm"
            className="btn-ghost w-8 h-8 p-0 hover:bg-gray-100 dark:hover:bg-gray-600"
            onClick={handleDecrease}
          >
            <Icon icon="ph:minus" className="text-xs" />
          </Button>
          <span className="w-8 text-center text-sm font-medium px-1">
            {item.quantity}
          </span>
          <Button
            size="sm"
            className="btn-ghost w-8 h-8 p-0 hover:bg-gray-100 dark:hover:bg-gray-600"
            onClick={handleIncrease}
          >
            <Icon icon="ph:plus" className="text-xs" />
          </Button>
        </div>
        
        <Button
          size="sm"
          className="btn-ghost w-8 h-8 p-0 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20"
          onClick={handleRemove}
          title="Remove item"
        >
          <Icon icon="ph:trash" className="text-xs" />
        </Button>
      </div>
    </div>
  );
};

const ShoppingCart = ({ 
  items = [], 
  onQuantityChange, 
  onRemove, 
  onClear, 
  isLoading = false 
}) => {
  const isEmpty = items.length === 0;

  return (
    <div className="space-y-3">
      {/* Cart Items */}
      <div className="space-y-2 max-h-80 overflow-y-auto">
        {items.map((item) => (
          <CartItem
            key={item.id}
            item={item}
            onQuantityChange={onQuantityChange}
            onRemove={onRemove}
          />
        ))}
        
        {isEmpty && !isLoading && (
          <div className="text-center py-12">
            <div className="w-16 h-16 mx-auto mb-4 bg-gray-100 dark:bg-gray-800 rounded-full flex items-center justify-center">
              <Icon icon="ph:shopping-cart" className="text-2xl text-gray-400" />
            </div>
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
              Cart is empty
            </h3>
            <p className="text-gray-500 dark:text-gray-400 text-sm">
              Add products to start a transaction
            </p>
          </div>
        )}

        {isLoading && (
          <div className="text-center py-8">
            <div className="animate-spin w-8 h-8 border-2 border-indigo-500 border-t-transparent rounded-full mx-auto mb-2"></div>
            <p className="text-sm text-gray-500 dark:text-gray-400">Loading cart...</p>
          </div>
        )}
      </div>

      {/* Cart Actions */}
      {!isEmpty && (
        <div className="border-t border-gray-200 dark:border-gray-700 pt-3">
          <Button
            icon="ph:trash"
            size="sm"
            className="btn-outline-danger w-full"
            onClick={onClear}
          >
            Clear Cart
          </Button>
        </div>
      )}
    </div>
  );
};

export default ShoppingCart;