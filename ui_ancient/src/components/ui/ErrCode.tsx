'use client'

import { Alert } from 'flowbite-react'
import { HiInformationCircle } from 'react-icons/hi'

type ErrCodeProps = {
  message?: string | undefined
}

export function ErrCode({ message }: ErrCodeProps) {
  return (
    <Alert color="warning" rounded icon={HiInformationCircle}>
      <span>
        <p>{message && message}</p>
      </span>
    </Alert>
  )
}
