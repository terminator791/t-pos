import React, { useState, useEffect } from "react";
import Card from "@/components/ui/Card";
import Icon from "@/components/ui/Icon";
import Button from "@/components/ui/Button";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import ErrorDisplay from "@/components/ui/ErrorDisplay";
import ShoppingCart from "@/components/pos/ShoppingCart";
import ProductGrid from "@/components/pos/ProductGrid";
import PaymentModal from "@/components/pos/PaymentModal";
import {
  useProducts,
  useCategories,
  useCarts,
  useAddToCart,
  useUpdateCartItem,
  useRemoveFromCart,
  useClearCart,
  useCreateTransaction,
  usePayTransaction,
  useShops,
  useProductByBarcode,
} from "@/services/api";
import { useSelector } from "react-redux";
import { toast } from "react-toastify";

const POSPage = () => {
  // State management
  const [selectedCategory, setSelectedCategory] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [barcodeInput, setBarcodeInput] = useState("");
  const [currentShopId, setCurrentShopId] = useState("");
  const [customerName, setCustomerName] = useState("");
  const [discountPercentage, setDiscountPercentage] = useState(0);
  const [discountAmount, setDiscountAmount] = useState(0);
  const [showPaymentModal, setShowPaymentModal] = useState(false);

  // Get user info from Redux
  const { user } = useSelector((state) => state.auth);

  // API hooks
  const { data: productsData, isLoading: productsLoading, error: productsError } = useProducts();
  const { data: categoriesData } = useCategories(currentShopId);
  const { data: cartsData, isLoading: cartsLoading, refetch: refetchCarts } = useCarts();
  const { data: shopsData } = useShops();
  
  // Mutations
  const addToCart = useAddToCart();
  const updateCartItem = useUpdateCartItem();
  const removeFromCart = useRemoveFromCart();
  const clearCart = useClearCart();
  const createTransaction = useCreateTransaction();
  const payTransaction = usePayTransaction();

  // Barcode search
  const { data: barcodeProduct, isLoading: barcodeLoading } = useProductByBarcode(barcodeInput, {
    enabled: !!barcodeInput && barcodeInput.length >= 3,
  });

  // Data processing
  const products = productsData?.data?.products || [];
  const categories = categoriesData?.data?.categories || [];
  const cartItems = cartsData?.data?.carts || [];
  const shops = shopsData?.data?.shops || [];

  // Set default shop if available
  useEffect(() => {
    if (!currentShopId && shops.length > 0) {
      setCurrentShopId(shops[0].id);
    }
  }, [shops, currentShopId]);

  // Filter products based on category and search
  const filteredProducts = products.filter((product) => {
    const matchesCategory = !selectedCategory || product.category_id === selectedCategory;
    const matchesSearch = !searchTerm || 
      product.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      product.barcode?.toLowerCase().includes(searchTerm.toLowerCase());
    return matchesCategory && matchesSearch && product.stock > 0;
  });

  // Calculate cart totals
  const subtotal = cartItems.reduce((sum, item) => sum + (item.quantity * item.unit_price), 0);
  const discountTotal = discountPercentage > 0 ? (subtotal * discountPercentage / 100) : discountAmount;
  const ppnAmount = subtotal * 0.11; // 11% PPN (Indonesian VAT)
  const total = subtotal - discountTotal + ppnAmount;

  // Handle adding product to cart
  const handleAddToCart = async (product) => {
    try {
      // Check if product already in cart
      const existingItem = cartItems.find(item => item.product_id === product.id);
      
      if (existingItem) {
        // Update quantity if already in cart
        await updateCartItem.mutateAsync({
          id: existingItem.id,
          quantity: existingItem.quantity + 1
        });
      } else {
        // Add new item to cart
        await addToCart.mutateAsync({
          shop_id: currentShopId,
          product_id: product.id,
          quantity: 1
        });
      }
    } catch (error) {
      console.error("Failed to add to cart:", error);
    }
  };

  // Handle quantity change
  const handleQuantityChange = async (cartItem, newQuantity) => {
    if (newQuantity <= 0) {
      await removeFromCart.mutateAsync(cartItem.id);
    } else {
      await updateCartItem.mutateAsync({
        id: cartItem.id,
        quantity: newQuantity
      });
    }
  };

  // Handle remove from cart
  const handleRemoveFromCart = async (cartItemId) => {
    await removeFromCart.mutateAsync(cartItemId);
  };

  // Handle clear cart
  const handleClearCart = async () => {
    await clearCart.mutateAsync();
  };

  // Handle barcode scan/input
  const handleBarcodeSubmit = () => {
    if (barcodeProduct?.data) {
      handleAddToCart(barcodeProduct.data);
      setBarcodeInput("");
    } else {
      toast.error("Product not found with this barcode");
    }
  };

  // Handle checkout
  const handleCheckout = () => {
    if (cartItems.length === 0) {
      toast.error("Cart is empty");
      return;
    }
    setShowPaymentModal(true);
  };

  // Handle payment processing
  const handlePayment = async (paymentData) => {
    try {
      const { amount, method, notes } = paymentData;

      // Create transaction
      const transactionData = {
        customer_name: customerName || "Walk-in Customer",
        items: cartItems.map(item => ({
          product_id: item.product_id,
          quantity: item.quantity
        })),
        cashier_name: user?.username || "Unknown",
        shop_id: currentShopId,
        discount: discountAmount,
        discount_percentage: discountPercentage,
      };

      const transactionResult = await createTransaction.mutateAsync(transactionData);
      
      if (transactionResult?.data?.id) {
        // Process payment
        await payTransaction.mutateAsync({
          transactionId: transactionResult.data.id,
          amount: amount
        });

        // Clear cart after successful transaction
        await clearCart.mutateAsync();
        
        // Reset form
        setCustomerName("");
        setDiscountPercentage(0);
        setDiscountAmount(0);
        setShowPaymentModal(false);
        
        toast.success("Transaction completed successfully!");
      }
    } catch (error) {
      console.error("Payment failed:", error);
    }
  };

  if (productsLoading || cartsLoading) {
    return <LoadingSpinner message="Loading POS system..." />;
  }

  if (productsError) {
    return <ErrorDisplay message="Failed to load products. Please try again." />;
  }

  return (
    <div className="space-y-5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            Point of Sale
          </h2>
          <p className="text-gray-500 dark:text-gray-400">
            Process sales and manage transactions
          </p>
        </div>
        
        {/* Shop Selector */}
        <div className="flex items-center space-x-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Current Shop
            </label>
            <select
              value={currentShopId}
              onChange={(e) => setCurrentShopId(e.target.value)}
              className="form-control"
            >
              <option value="">Select Shop</option>
              {shops.map((shop) => (
                <option key={shop.id} value={shop.id}>
                  {shop.name}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Products Section */}
        <div className="lg:col-span-2 space-y-5">
          {/* Search and Filters */}
          <Card>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {/* Search */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Search Products
                </label>
                <div className="flex">
                  <input
                    type="text"
                    placeholder="Search by name..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="form-control rounded-r-none"
                  />
                  <Button
                    icon="ph:magnifying-glass"
                    className="btn-primary rounded-l-none"
                  />
                </div>
              </div>

              {/* Category Filter */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Category
                </label>
                <select
                  value={selectedCategory}
                  onChange={(e) => setSelectedCategory(e.target.value)}
                  className="form-control"
                >
                  <option value="">All Categories</option>
                  {categories.map((category) => (
                    <option key={category.id} value={category.id}>
                      {category.name}
                    </option>
                  ))}
                </select>
              </div>

              {/* Barcode Scanner */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Barcode Scanner
                </label>
                <div className="flex">
                  <input
                    type="text"
                    placeholder="Scan or enter barcode..."
                    value={barcodeInput}
                    onChange={(e) => setBarcodeInput(e.target.value)}
                    onKeyPress={(e) => e.key === "Enter" && handleBarcodeSubmit()}
                    className="form-control rounded-r-none"
                  />
                  <Button
                    icon="ph:barcode"
                    className="btn-secondary rounded-l-none"
                    onClick={handleBarcodeSubmit}
                    isLoading={barcodeLoading}
                  />
                </div>
              </div>
            </div>
          </Card>

          {/* Products Grid */}
          <Card title="Products">
            <ProductGrid
              products={filteredProducts}
              onAddToCart={handleAddToCart}
              isLoading={productsLoading}
              searchTerm={searchTerm}
              selectedCategory={selectedCategory}
            />
          </Card>
        </div>

        {/* Cart Section */}
        <div className="space-y-5">
          {/* Customer Information */}
          <Card title="Customer Information">
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Customer Name
                </label>
                <input
                  type="text"
                  placeholder="Enter customer name (optional)"
                  value={customerName}
                  onChange={(e) => setCustomerName(e.target.value)}
                  className="form-control"
                />
              </div>
            </div>
          </Card>

          {/* Shopping Cart */}
          <Card title="Shopping Cart">
            <ShoppingCart
              items={cartItems}
              onQuantityChange={handleQuantityChange}
              onRemove={handleRemoveFromCart}
              onClear={handleClearCart}
              isLoading={cartsLoading}
            />
          </Card>

          {/* Discounts */}
          <Card title="Discounts">
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Discount Percentage (%)
                </label>
                <input
                  type="number"
                  min="0"
                  max="100"
                  step="0.01"
                  value={discountPercentage}
                  onChange={(e) => {
                    setDiscountPercentage(Number(e.target.value));
                    setDiscountAmount(0); // Clear amount discount when percentage is used
                  }}
                  className="form-control"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Discount Amount ($)
                </label>
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  value={discountAmount}
                  onChange={(e) => {
                    setDiscountAmount(Number(e.target.value));
                    setDiscountPercentage(0); // Clear percentage discount when amount is used
                  }}
                  className="form-control"
                />
              </div>
            </div>
          </Card>

          {/* Order Summary */}
          <Card title="Order Summary">
            <div className="space-y-3">
              <div className="flex justify-between">
                <span className="text-gray-600 dark:text-gray-400">Subtotal:</span>
                <span className="font-medium">${subtotal.toFixed(2)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600 dark:text-gray-400">Discount:</span>
                <span className="font-medium text-red-600">-${discountTotal.toFixed(2)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600 dark:text-gray-400">PPN (11%):</span>
                <span className="font-medium">${ppnAmount.toFixed(2)}</span>
              </div>
              <div className="border-t pt-3">
                <div className="flex justify-between text-lg font-bold">
                  <span>Total:</span>
                  <span>${total.toFixed(2)}</span>
                </div>
              </div>
            </div>
            
            <Button
              className="btn-primary w-full mt-4"
              onClick={handleCheckout}
              disabled={cartItems.length === 0}
              icon="ph:credit-card"
            >
              Checkout
            </Button>
          </Card>
        </div>
      </div>

      {/* Payment Modal */}
      <PaymentModal
        isOpen={showPaymentModal}
        onClose={() => setShowPaymentModal(false)}
        onConfirmPayment={handlePayment}
        orderSummary={{
          subtotal,
          discount: discountTotal,
          tax: ppnAmount,
          total
        }}
        isProcessing={createTransaction.isPending || payTransaction.isPending}
      />
    </div>
  );
};

export default POSPage;