import React, { useState } from "react";
import Card from "@/components/ui/Card";
import Icon from "@/components/ui/Icon";
import Button from "@/components/ui/Button";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import ErrorDisplay from "@/components/ui/ErrorDisplay";
import Modal from "@/components/ui/Modal";
import {
  useTransactionHistories,
  useTransaction,
  useTransactionProducts,
  useShops,
  usePayments,
} from "@/services/api";

const TransactionHistoriesPage = () => {
  const [selectedShop, setSelectedShop] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [dateFilter, setDateFilter] = useState("");
  const [selectedTransaction, setSelectedTransaction] = useState(null);
  const [showDetailsModal, setShowDetailsModal] = useState(false);

  // API hooks
  const { data: historiesData, isLoading, error, refetch } = useTransactionHistories(selectedShop);
  const { data: shopsData } = useShops();
  const { data: transactionData, isLoading: isLoadingTransaction } = useTransaction(
    selectedTransaction?.transaction_id
  );
  const { data: transactionProductsData } = useTransactionProducts(
    selectedTransaction?.transaction_id
  );
  const { data: paymentsData } = usePayments(selectedShop);

  const histories = historiesData?.data?.histories || [];
  const totalHistories = historiesData?.data?.count || 0;
  const shops = shopsData?.data?.shops || [];
  const transaction = transactionData?.data?.transaction;
  const transactionProducts = transactionProductsData?.data?.transaction_products || [];
  const payments = paymentsData?.data?.payments || [];

  // Filter histories based on search term and status
  const filteredHistories = histories.filter((history) => {
    const matchesSearch = 
      !searchTerm ||
      history.transaction_id?.toString().includes(searchTerm) ||
      history.id?.toString().includes(searchTerm);
    const matchesStatus = !statusFilter || history.status === statusFilter;
    const matchesDate = !dateFilter || (
      history.created_at && 
      new Date(history.created_at).toDateString() === new Date(dateFilter).toDateString()
    );
    return matchesSearch && matchesStatus && matchesDate;
  });

  const handleViewDetails = (history) => {
    setSelectedTransaction(history);
    setShowDetailsModal(true);
  };

  const getShopName = (shopId) => {
    const shop = shops.find(s => s.id === shopId);
    return shop?.name || "Unknown Shop";
  };

  const getStatusBadge = (status) => {
    const statusConfig = {
      completed: { color: "success", label: "Completed" },
      pending: { color: "warning", label: "Pending" },
      cancelled: { color: "danger", label: "Cancelled" },
      processing: { color: "info", label: "Processing" },
    };
    
    const config = statusConfig[status] || { color: "secondary", label: status };
    
    return (
      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-${config.color}-100 text-${config.color}-800 dark:bg-${config.color}-900/20 dark:text-${config.color}-400`}>
        {config.label}
      </span>
    );
  };

  const formatCurrency = (amount) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
    }).format(amount || 0);
  };

  return (
    <div className="space-y-5">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            Transaction Histories
          </h2>
          <p className="text-gray-500 dark:text-gray-400">
            View detailed transaction histories and records
          </p>
        </div>
        <Button
          icon="ph:download"
          className="btn-outline-primary"
          onClick={() => {/* TODO: Implement export */}}
        >
          Export Data
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-5">
        <Card>
          <div className="flex items-center space-x-3">
            <div className="flex-none">
              <div className="w-12 h-12 rounded-full bg-primary-500/10 flex items-center justify-center">
                <Icon
                  icon="ph:clock"
                  className="text-2xl text-primary-500"
                />
              </div>
            </div>
            <div className="flex-1">
              <div className="text-slate-900 dark:text-white text-lg font-medium">
                {totalHistories}
              </div>
              <div className="text-slate-500 dark:text-slate-400 text-sm">
                Total Records
              </div>
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center space-x-3">
            <div className="flex-none">
              <div className="w-12 h-12 rounded-full bg-success-500/10 flex items-center justify-center">
                <Icon
                  icon="ph:check-circle"
                  className="text-2xl text-success-500"
                />
              </div>
            </div>
            <div className="flex-1">
              <div className="text-slate-900 dark:text-white text-lg font-medium">
                {filteredHistories.filter(h => h.status === 'completed').length}
              </div>
              <div className="text-slate-500 dark:text-slate-400 text-sm">
                Completed
              </div>
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center space-x-3">
            <div className="flex-none">
              <div className="w-12 h-12 rounded-full bg-warning-500/10 flex items-center justify-center">
                <Icon
                  icon="ph:hourglass"
                  className="text-2xl text-warning-500"
                />
              </div>
            </div>
            <div className="flex-1">
              <div className="text-slate-900 dark:text-white text-lg font-medium">
                {filteredHistories.filter(h => h.status === 'pending').length}
              </div>
              <div className="text-slate-500 dark:text-slate-400 text-sm">
                Pending
              </div>
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center space-x-3">
            <div className="flex-none">
              <div className="w-12 h-12 rounded-full bg-info-500/10 flex items-center justify-center">
                <Icon
                  icon="ph:list-dashes"
                  className="text-2xl text-info-500"
                />
              </div>
            </div>
            <div className="flex-1">
              <div className="text-slate-900 dark:text-white text-lg font-medium">
                {filteredHistories.length}
              </div>
              <div className="text-slate-500 dark:text-slate-400 text-sm">
                Filtered Results
              </div>
            </div>
          </div>
        </Card>
      </div>

      {/* Filters */}
      <Card>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Search
            </label>
            <input
              type="text"
              placeholder="Search by transaction ID..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="form-control"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Shop
            </label>
            <select
              value={selectedShop}
              onChange={(e) => setSelectedShop(e.target.value)}
              className="form-control"
            >
              <option value="">All Shops</option>
              {shops.map((shop) => (
                <option key={shop.id} value={shop.id}>
                  {shop.name}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Status
            </label>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="form-control"
            >
              <option value="">All Status</option>
              <option value="completed">Completed</option>
              <option value="pending">Pending</option>
              <option value="cancelled">Cancelled</option>
              <option value="processing">Processing</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Date
            </label>
            <input
              type="date"
              value={dateFilter}
              onChange={(e) => setDateFilter(e.target.value)}
              className="form-control"
            />
          </div>
        </div>
      </Card>

      {/* Transaction Histories Table */}
      <Card>
        {isLoading && (
          <div className="text-center py-8">
            <LoadingSpinner />
          </div>
        )}

        {error && (
          <div className="text-center py-8">
            <ErrorDisplay
              message="Failed to load transaction histories."
              onRetry={refetch}
            />
          </div>
        )}

        {!isLoading && !error && (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-gray-800">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Transaction
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Shop
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Date
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
                {filteredHistories.length === 0 ? (
                  <tr>
                    <td
                      colSpan="5"
                      className="px-6 py-12 text-center text-gray-500 dark:text-gray-400"
                    >
                      <Icon
                        icon="ph:clock"
                        className="mx-auto h-12 w-12 text-gray-400 mb-4"
                      />
                      <h3 className="text-sm font-medium text-gray-900 dark:text-white">
                        No transaction histories found
                      </h3>
                      <p className="text-sm text-gray-500 dark:text-gray-400">
                        Transaction histories will appear here once transactions are created.
                      </p>
                    </td>
                  </tr>
                ) : (
                  filteredHistories.map((history) => (
                    <tr
                      key={history.id}
                      className="hover:bg-gray-50 dark:hover:bg-gray-800/50"
                    >
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <div className="flex-shrink-0 h-10 w-10">
                            <div className="h-10 w-10 rounded-lg bg-primary-100 dark:bg-primary-900/20 flex items-center justify-center">
                              <Icon
                                icon="ph:receipt"
                                className="h-5 w-5 text-primary-600 dark:text-primary-400"
                              />
                            </div>
                          </div>
                          <div className="ml-4">
                            <div className="text-sm font-medium text-gray-900 dark:text-white">
                              Transaction #{history.transaction_id}
                            </div>
                            <div className="text-sm text-gray-500 dark:text-gray-400">
                              History ID: {history.id}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                        {getShopName(history.shop_id)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {getStatusBadge(history.status)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                        {history.created_at
                          ? new Date(history.created_at).toLocaleString()
                          : "N/A"}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <Button
                          icon="ph:eye"
                          size="sm"
                          className="btn-outline-primary"
                          onClick={() => handleViewDetails(history)}
                        >
                          View Details
                        </Button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* Transaction Details Modal */}
      <Modal
        title="Transaction Details"
        activeModal={showDetailsModal}
        onClose={() => setShowDetailsModal(false)}
        className="max-w-4xl"
      >
        {isLoadingTransaction ? (
          <div className="text-center py-8">
            <LoadingSpinner />
          </div>
        ) : transaction ? (
          <div className="space-y-6">
            {/* Transaction Info */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <Card title="Transaction Information">
                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-gray-500">Transaction ID:</span>
                    <span className="font-medium">#{transaction.id}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-500">Shop:</span>
                    <span className="font-medium">{getShopName(transaction.shop_id)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-500">Total Amount:</span>
                    <span className="font-medium">{formatCurrency(transaction.total_price)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-500">Discount:</span>
                    <span className="font-medium">{transaction.discount_percentage || 0}%</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-500">Status:</span>
                    {getStatusBadge(transaction.payment_status)}
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-500">Date:</span>
                    <span className="font-medium">
                      {new Date(transaction.created_at).toLocaleString()}
                    </span>
                  </div>
                </div>
              </Card>

              <Card title="Payment Information">
                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-gray-500">Payment Method:</span>
                    <span className="font-medium">{transaction.payment_method || "Cash"}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-500">Amount Paid:</span>
                    <span className="font-medium">{formatCurrency(transaction.total_price)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-500">Change:</span>
                    <span className="font-medium">{formatCurrency(transaction.change || 0)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-500">Profit:</span>
                    <span className="font-medium text-success-600">
                      {formatCurrency(transaction.profit_transaction || 0)}
                    </span>
                  </div>
                </div>
              </Card>
            </div>

            {/* Transaction Items */}
            {transactionProducts.length > 0 && (
              <Card title="Transaction Items">
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                    <thead className="bg-gray-50 dark:bg-gray-800">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Product
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Quantity
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Unit Price
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Total
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
                      {transactionProducts.map((item) => (
                        <tr key={item.id}>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                            {item.product_name || `Product ID: ${item.product_id}`}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                            {item.quantity}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                            {formatCurrency(item.unit_price)}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                            {formatCurrency(item.total_price)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </Card>
            )}
          </div>
        ) : (
          <div className="text-center py-8">
            <Icon
              icon="ph:warning"
              className="mx-auto h-12 w-12 text-yellow-400 mb-4"
            />
            <h3 className="text-sm font-medium text-gray-900 dark:text-white">
              Transaction not found
            </h3>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              The transaction details could not be loaded.
            </p>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default TransactionHistoriesPage;