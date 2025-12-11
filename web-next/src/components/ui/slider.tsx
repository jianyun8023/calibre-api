"use client"

import * as React from "react"
import { cn } from "@/lib/utils"

interface SliderProps {
  value: number[]
  onValueChange: (value: number[]) => void
  min?: number
  max?: number
  step?: number
  className?: string
}

const Slider = React.forwardRef<HTMLDivElement, SliderProps>(
  ({ value, onValueChange, min = 0, max = 100, step = 1, className }, ref) => {
    const handleChange = (index: number, newValue: number) => {
      const newValues = [...value]
      newValues[index] = Math.max(min, Math.min(max, newValue))
      
      // Ensure proper order for range sliders
      if (value.length === 2) {
        if (index === 0 && newValues[0] > newValues[1]) {
          newValues[1] = newValues[0]
        } else if (index === 1 && newValues[1] < newValues[0]) {
          newValues[0] = newValues[1]
        }
      }
      
      onValueChange(newValues)
    }

    const getPercentage = (val: number) => ((val - min) / (max - min)) * 100

    return (
      <div
        ref={ref}
        className={cn("relative flex w-full touch-none select-none items-center", className)}
      >
        <div className="relative h-2 w-full grow overflow-hidden rounded-full bg-secondary">
          {/* Range highlight for dual sliders */}
          {value.length === 2 && (
            <div
              className="absolute h-full bg-primary"
              style={{
                left: `${getPercentage(value[0])}%`,
                width: `${getPercentage(value[1]) - getPercentage(value[0])}%`,
              }}
            />
          )}
          
          {/* Single value highlight */}
          {value.length === 1 && (
            <div
              className="absolute h-full bg-primary"
              style={{
                width: `${getPercentage(value[0])}%`,
              }}
            />
          )}
          
          {/* Thumbs */}
          {value.map((val, index) => (
            <input
              key={index}
              type="range"
              min={min}
              max={max}
              step={step}
              value={val}
              onChange={(e) => handleChange(index, Number(e.target.value))}
              className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
              style={{
                zIndex: 10 + index,
              }}
            />
          ))}
          
          {/* Visual thumbs */}
          {value.map((val, index) => (
            <div
              key={`thumb-${index}`}
              className="absolute block h-5 w-5 rounded-full border-2 border-primary bg-background ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 -translate-x-1/2 -translate-y-1/2 top-1/2"
              style={{
                left: `${getPercentage(val)}%`,
                pointerEvents: 'none',
              }}
            />
          ))}
        </div>
      </div>
    )
  }
)

Slider.displayName = "Slider"

export { Slider }