'use client'

import { Button } from 'flowbite-react'
import { HiChevronLeft, HiChevronRight } from 'react-icons/hi'

type SurveyFooterProps = {
  current: number
  size: number
  canGoBack: boolean
  canGoForward: boolean
  onBackClick: () => void
  onForwardClick: () => void
}

export default function SurveyFooter({
  current,
  size,
  canGoBack,
  canGoForward,
  onBackClick,
  onForwardClick,
}: SurveyFooterProps) {
  if (current === 1) {
    canGoBack = false
  }

  if (current === size) {
    canGoForward = false
  }

  return (
    <div className="footer">
      <div className="h-full max-w-lg mx-auto">
        <div className="flex items-center justify-center">
          <div className="navi">
            <Button
              type="button"
              disabled={!canGoBack}
              className="prev"
              onClick={onBackClick}
            >
              <HiChevronLeft />
            </Button>
            <span className="curr">
              {current} of {size}
            </span>
            <Button
              type="button"
              disabled={!canGoForward}
              className="next"
              onClick={onForwardClick}
            >
              <HiChevronRight />
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
