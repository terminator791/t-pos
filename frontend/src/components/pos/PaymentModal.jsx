import React, { useState, useEffect } from "react";
import Modal from "@/components/ui/Modal";
import Button from "@/components/ui/Button";
import Icon from "@/components/ui/Icon";

const PaymentModal = ({
  isOpen,
  onClose,
  onConfirmPayment,
  orderSummary,
  isProcessing = false
}) => {
  const [paymentAmount, setPaymentAmount] = useState(0);
  const [paymentMethod, setPaymentMethod] = useState("cash");
  const [notes, setNotes] = useState("");

  const { subtotal = 0, discount = 0, tax = 0, total = 0 } = orderSummary || {};
  const change = paymentAmount - total;
  const isValidPayment = paymentAmount >= total;

  // Set default payment amount when modal opens
  useEffect(() => {
    if (isOpen && total > 0) {
      setPaymentAmount(total);
    }
  }, [isOpen, total]);

  const handlePaymentMethodChange = (method) => {
    setPaymentMethod(method);
    // For non-cash payments, set exact amount
    if (method !== "cash") {
      setPaymentAmount(total);
    }
  };

  const handleQuickAmountSelect = (amount) => {
    setPaymentAmount(amount);
  };

  const handleConfirm = () => {
    if (!isValidPayment) return;

    onConfirmPayment({
      amount: paymentAmount,
      method: paymentMethod,
      change: change,
      notes: notes.trim()
    });
  };

  const handleClose = () => {
    if (!isProcessing) {
      setPaymentAmount(0);
      setPaymentMethod("cash");
      setNotes("");
      onClose();
    }
  };

  // Quick amount buttons for cash payments
  const quickAmounts = [
    total,
    Math.ceil(total / 50) * 50, // Round to nearest 50
    Math.ceil(total / 100) * 100, // Round to nearest 100
    total + 50,
    total + 100
  ].filter((amount, index, arr) => arr.indexOf(amount) === index && amount >= total);

  const paymentMethods = [
    { id: "cash", label: "Cash", icon: "ph:money", color: "green" },
    { id: "card", label: "Credit/Debit Card", icon: "ph:credit-card", color: "blue" },
    { id: "digital", label: "Digital Payment", icon: "ph:device-mobile", color: "purple" },
    { id: "bank_transfer", label: "Bank Transfer", icon: "ph:bank", color: "indigo" }
  ];

  return (
    <Modal
      title="Process Payment"
      activeModal={isOpen}
      onClose={handleClose}
      className="max-w-lg"
      noCloseButton={isProcessing}
    >
      <div className="space-y-6">
        {/* Order Summary */}
        <div className="bg-gray-50 dark:bg-gray-800 p-4 rounded-lg">
          <h3 className="font-medium text-gray-900 dark:text-white mb-3">Order Summary</h3>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-600 dark:text-gray-400">Subtotal:</span>
              <span className="font-medium">${subtotal.toFixed(2)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600 dark:text-gray-400">Discount:</span>
              <span className="font-medium text-red-600">-${discount.toFixed(2)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600 dark:text-gray-400">Tax (PPN):</span>
              <span className="font-medium">${tax.toFixed(2)}</span>
            </div>
            <div className="border-t border-gray-200 dark:border-gray-700 pt-2">
              <div className="flex justify-between text-lg font-bold">
                <span>Total:</span>
                <span className="text-indigo-600 dark:text-indigo-400">${total.toFixed(2)}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Payment Methods */}
        <div>
          <h3 className="font-medium text-gray-900 dark:text-white mb-3">Payment Method</h3>
          <div className="grid grid-cols-2 gap-3">
            {paymentMethods.map((method) => (
              <button
                key={method.id}
                onClick={() => handlePaymentMethodChange(method.id)}
                className={`
                  p-3 rounded-lg border-2 transition-all duration-200 flex items-center space-x-2
                  ${paymentMethod === method.id
                    ? `border-${method.color}-500 bg-${method.color}-50 dark:bg-${method.color}-900/20`
                    : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
                  }
                `}
                disabled={isProcessing}
              >
                <Icon 
                  icon={method.icon} 
                  className={`text-lg ${
                    paymentMethod === method.id 
                      ? `text-${method.color}-600 dark:text-${method.color}-400` 
                      : 'text-gray-500'
                  }`} 
                />
                <span className={`text-sm font-medium ${
                  paymentMethod === method.id 
                    ? `text-${method.color}-800 dark:text-${method.color}-300` 
                    : 'text-gray-700 dark:text-gray-300'
                }`}>
                  {method.label}
                </span>
              </button>
            ))}
          </div>
        </div>

        {/* Payment Amount */}
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Amount Received
          </label>
          <div className="relative">
            <span className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-500 dark:text-gray-400">
              $
            </span>
            <input
              type="number"
              min={total}
              step="0.01"
              value={paymentAmount}
              onChange={(e) => setPaymentAmount(Number(e.target.value))}
              className="form-control pl-8"
              placeholder="Enter payment amount"
              disabled={isProcessing || paymentMethod !== "cash"}
            />
          </div>
          
          {/* Quick Amount Buttons for Cash */}
          {paymentMethod === "cash" && (
            <div className="mt-3">
              <div className="flex flex-wrap gap-2">
                {quickAmounts.map((amount) => (
                  <Button
                    key={amount}
                    size="sm"
                    className="btn-outline-secondary"
                    onClick={() => handleQuickAmountSelect(amount)}
                    disabled={isProcessing}
                  >
                    ${amount.toFixed(0)}
                  </Button>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Change Display */}
        {paymentMethod === "cash" && paymentAmount >= total && (
          <div className="bg-green-50 dark:bg-green-900/20 p-4 rounded-lg border border-green-200 dark:border-green-800">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <Icon icon="ph:money" className="text-green-600 dark:text-green-400" />
                <span className="font-medium text-green-800 dark:text-green-300">Change:</span>
              </div>
              <span className="text-xl font-bold text-green-800 dark:text-green-300">
                ${change.toFixed(2)}
              </span>
            </div>
          </div>
        )}

        {/* Payment Notes */}
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Notes (Optional)
          </label>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            className="form-control"
            rows={2}
            placeholder="Add any notes about this payment..."
            disabled={isProcessing}
          />
        </div>

        {/* Error Message */}
        {!isValidPayment && paymentAmount > 0 && (
          <div className="bg-red-50 dark:bg-red-900/20 p-3 rounded-lg border border-red-200 dark:border-red-800">
            <div className="flex items-center space-x-2">
              <Icon icon="ph:warning" className="text-red-600 dark:text-red-400" />
              <span className="text-sm text-red-800 dark:text-red-300">
                Payment amount cannot be less than the total amount.
              </span>
            </div>
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex space-x-3 pt-4 border-t border-gray-200 dark:border-gray-700">
          <Button
            className="btn-outline-secondary flex-1"
            onClick={handleClose}
            disabled={isProcessing}
          >
            Cancel
          </Button>
          <Button
            className="btn-primary flex-1"
            onClick={handleConfirm}
            disabled={!isValidPayment || isProcessing}
            isLoading={isProcessing}
            icon="ph:check"
          >
            {paymentMethod === "cash" ? "Complete Payment" : "Process Payment"}
          </Button>
        </div>
      </div>
    </Modal>
  );
};

export default PaymentModal;