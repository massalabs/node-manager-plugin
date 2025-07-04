import React from 'react';

import {
  PopupModal,
  PopupModalContent,
  PopupModalHeader,
  PopupModalFooter,
} from '@massalabs/react-ui-kit';

interface ConfirmModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  children: React.ReactNode;
}

const ConfirmModal: React.FC<ConfirmModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  title,
  children,
}) => {
  if (!isOpen) {
    return null;
  }

  return (
    <PopupModal
      fullMode={true}
      customClass="border-2 border-black bg-gray-850 w-full md:w-1/2 lg:w-1/3"
      onClose={onClose}
    >
      <PopupModalHeader customClassHeader="bg-gray-900">
        <p className="mas-title mb-6">{title}</p>
      </PopupModalHeader>

      <PopupModalContent customClassContent="bg-gray-850 pb-5 pt-5">
        {children}
      </PopupModalContent>

      <PopupModalFooter customClassFooter="bg-gray-850 pt-5 pb-1">
        <div className="flex justify-between w-full">
          <button
            className="bg-gray-500 text-white font-bold px-4 py-2 rounded hover:bg-gray-600"
            onClick={onClose}
          >
            Cancel
          </button>
          <button
            className="bg-green-500 text-white font-bold px-4 py-2 rounded hover:bg-green-600"
            onClick={onConfirm}
          >
            Confirm
          </button>
        </div>
      </PopupModalFooter>
    </PopupModal>
  );
};

export default ConfirmModal;
