import * as React from "react"
import { cn } from "@/lib/utils"

const Alert = React.forwardRef(({ className, variant = "default", ...props }, ref) => {
  const variants = {
    default: "bg-white text-slate-950 border-slate-200",
    destructive: "border-red-500/50 text-red-600 bg-red-50",
    success: "border-green-500/50 text-green-600 bg-green-50",
  }
  return (
    <div
      ref={ref}
      role="alert"
      className={cn("relative w-full rounded-lg border p-4", variants[variant], className)}
      {...props}
    />
  )
})
Alert.displayName = "Alert"

const AlertDescription = React.forwardRef(({ className, ...props }, ref) => (
  <div ref={ref} className={cn("text-sm [&_p]:leading-relaxed", className)} {...props} />
))
AlertDescription.displayName = "AlertDescription"

export { Alert, AlertDescription }
