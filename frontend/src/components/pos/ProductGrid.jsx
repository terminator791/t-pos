import React from "react";
import Icon from "@/components/ui/Icon";
import Button from "@/components/ui/Button";

const ProductCard = ({ product, onAddToCart, isOutOfStock = false }) => {
  const handleClick = () => {
    if (!isOutOfStock) {
      onAddToCart(product);
    }
  };

  return (
    <div
      className={`
        bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 
        rounded-lg p-4 transition-all duration-200 cursor-pointer
        ${isOutOfStock 
          ? 'opacity-50 cursor-not-allowed' 
          : 'hover:shadow-md hover:border-indigo-300 dark:hover:border-indigo-600 hover:scale-105'
        }
      `}
      onClick={handleClick}
    >
      {/* Product Image */}
      <div className="aspect-square bg-gray-100 dark:bg-gray-700 rounded-lg flex items-center justify-center mb-3 relative overflow-hidden">
        {product.photo ? (
          <img
            src={product.photo}
            alt={product.name}
            className="w-full h-full object-cover rounded-lg"
            onError={(e) => {
              e.target.style.display = 'none';
              e.target.nextSibling.style.display = 'flex';
            }}
          />
        ) : null}
        <div 
          className={`w-full h-full flex items-center justify-center ${
            product.photo ? 'hidden' : ''
          }`}
        >
          <Icon icon="ph:package" className="text-4xl text-gray-400" />
        </div>
        
        {/* Stock Badge */}
        {isOutOfStock && (
          <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center">
            <span className="bg-red-500 text-white px-2 py-1 rounded text-xs font-medium">
              Out of Stock
            </span>
          </div>
        )}
        
        {/* Low Stock Warning */}
        {!isOutOfStock && product.stock <= 10 && (
          <div className="absolute top-2 right-2">
            <span className="bg-yellow-500 text-white px-2 py-1 rounded text-xs font-medium">
              Low Stock
            </span>
          </div>
        )}
      </div>

      {/* Product Info */}
      <div className="space-y-1">
        <h3 className="font-medium text-gray-900 dark:text-white text-sm truncate">
          {product.name}
        </h3>
        
        <div className="flex items-center justify-between">
          <p className="text-gray-500 dark:text-gray-400 text-xs">
            Stock: {product.stock} {product.unit}
          </p>
          {product.barcode && (
            <p className="text-gray-400 text-xs font-mono">
              #{product.barcode.slice(-4)}
            </p>
          )}
        </div>
        
        <div className="flex items-center justify-between">
          <p className="text-lg font-bold text-indigo-600 dark:text-indigo-400">
            ${product.sale?.toFixed(2)}
          </p>
          
          {!isOutOfStock && (
            <Button
              size="sm"
              className="btn-primary p-1"
              onClick={(e) => {
                e.stopPropagation();
                handleClick();
              }}
            >
              <Icon icon="ph:plus" className="text-sm" />
            </Button>
          )}
        </div>
        
        {/* Category */}
        {product.category && (
          <p className="text-xs text-gray-400 truncate">
            {product.category.name}
          </p>
        )}
      </div>
    </div>
  );
};

const ProductGrid = ({ 
  products = [], 
  onAddToCart, 
  isLoading = false,
  searchTerm = "",
  selectedCategory = ""
}) => {
  if (isLoading) {
    return (
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {Array.from({ length: 8 }).map((_, index) => (
          <div
            key={index}
            className="bg-gray-200 dark:bg-gray-700 rounded-lg p-4 animate-pulse"
          >
            <div className="aspect-square bg-gray-300 dark:bg-gray-600 rounded-lg mb-3"></div>
            <div className="space-y-2">
              <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded"></div>
              <div className="h-3 bg-gray-300 dark:bg-gray-600 rounded w-3/4"></div>
              <div className="h-5 bg-gray-300 dark:bg-gray-600 rounded w-1/2"></div>
            </div>
          </div>
        ))}
      </div>
    );
  }

  const hasFilters = searchTerm || selectedCategory;
  const filterText = hasFilters ? 
    `Showing products${searchTerm ? ` matching "${searchTerm}"` : ''}${selectedCategory ? ` in selected category` : ''}` :
    'All products';

  return (
    <div className="space-y-4">
      {/* Results Header */}
      <div className="flex items-center justify-between">
        <p className="text-sm text-gray-600 dark:text-gray-400">
          {filterText} ({products.length} items)
        </p>
        
        {hasFilters && (
          <Button
            size="sm"
            className="btn-outline-secondary"
            onClick={() => window.location.reload()} // Simple reset for now
          >
            Clear Filters
          </Button>
        )}
      </div>

      {/* Products Grid */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {products.map((product) => (
          <ProductCard
            key={product.id}
            product={product}
            onAddToCart={onAddToCart}
            isOutOfStock={product.stock <= 0}
          />
        ))}
      </div>
      
      {/* Empty State */}
      {products.length === 0 && !isLoading && (
        <div className="text-center py-12">
          <div className="w-16 h-16 mx-auto mb-4 bg-gray-100 dark:bg-gray-800 rounded-full flex items-center justify-center">
            <Icon icon="ph:package" className="text-2xl text-gray-400" />
          </div>
          <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
            No products found
          </h3>
          <p className="text-gray-500 dark:text-gray-400">
            {hasFilters 
              ? "Try adjusting your search criteria or filters."
              : "No products are available at the moment."
            }
          </p>
        </div>
      )}
    </div>
  );
};

export default ProductGrid;