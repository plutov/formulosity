import AppLayout from 'components/app/AppLayout'

export default async function NotFoundPage() {
  return (
    <AppLayout>
      <div className="mt-8 flex flex-col items-center gap-4">
        <h1 className="md:display h2 w-full px-4 text-center md:w-[802px] md:px-0">
          404
        </h1>
        <p className="body-xl px-4 text-center text-slate-11 md:w-[572px] md:px-0">
          Page not found :(
        </p>
      </div>
    </AppLayout>
  )
}
