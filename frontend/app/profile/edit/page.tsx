"use client";
import { Form } from "@nextui-org/form";
import { useEffect, useState } from "react";
import { Input } from "@nextui-org/input";
import { DatePicker } from "@nextui-org/date-picker";
import { Button } from "@nextui-org/button";
import { CalendarDate } from "@internationalized/date";
import { toast } from "react-toastify";

import { PageHeader } from "@/components/page-header";
import { WorkoutUser } from "@/api/api.generated";
import { authApi } from "@/api/api";
import { Loading } from "@/components/loading";
import Avatar from "@/components/avatar";

function DataForm({ user }: { user: WorkoutUser }) {
  function protoToCalendarDate(s: string) {
    if (!s) {
      return null;
    }

    const date = new Date(s);

    return new CalendarDate(
      date.getUTCFullYear(),
      date.getUTCMonth() + 1,
      date.getUTCDate(),
    );
  }

  const [userDateOfBirth, setUserDateOfBirth] = useState(
    protoToCalendarDate(user.dateOfBirth!),
  );
  const [userFirstName, setUserFirstName] = useState(user.firstName);
  const [userLastName, setUserLastName] = useState(user.lastName);
  const [userWeight, setUserWeight] = useState(user.weight!);
  const [userHeight, setUserHeight] = useState(user.height!);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    await authApi.v1
      .userServiceUpdateUser({
        firstName: userFirstName ? userFirstName : undefined,
        lastName: userLastName ? userLastName : undefined,
        dateOfBirth: userDateOfBirth
          ? userDateOfBirth.toDate("UTC").toISOString()
          : undefined,
        weight: userWeight > 0 ? userWeight : undefined,
        height: userHeight > 0 ? userHeight : undefined,
      })
      .then(() => {
        toast.success("Данные успешно обновлены");
      })
      .catch((error) => {
        console.log(error);
        toast.error("Ошибка при обновлении данных");
      });
  }

  return (
    <Form className="flex flex-col pt-4 px-4 gap-4" onSubmit={handleSubmit}>
      <Input
        label="Имя"
        placeholder="Имя"
        value={userFirstName}
        onChange={(e) => setUserFirstName(e.target.value)}
      />
      <Input
        label="Фамилия"
        placeholder="Фамилия"
        value={userLastName}
        onChange={(e) => setUserLastName(e.target.value)}
      />
      <DatePicker
        label="Дата рождения"
        value={userDateOfBirth}
        onChange={setUserDateOfBirth}
      />
      <div className="flex flex-row gap-4 w-full">
        <Input
          className="w-1/2"
          label="Вес"
          type="number"
          value={String(userWeight)}
          onValueChange={(value) => setUserWeight(Number(value))}
        />
        <Input
          className="w-1/2"
          label="Рост"
          type="number"
          value={String(userHeight)}
          onValueChange={(value) => setUserHeight(Number(value))}
        />
      </div>
      <Button className="self-end w-full" color="success" type="submit">
        Сохранить
      </Button>
    </Form>
  );
}

function AvatarSection({
  src,
  setProfilePictureURL,
}: {
  src: string;
  setProfilePictureURL: (url: string) => void;
}) {
  async function handleUploadAvatar(e: React.ChangeEvent<HTMLInputElement>) {
    if (!e.target.files || e.target.files.length === 0) {
      return;
    }

    const file = e.target.files[0];

    const filename =
      (Math.random() + 1).toString(36).substring(2) + "-" + file.name;

    const data = await authApi.v1
      .fileServicePresignUpload({
        filename: filename,
        contentType: file.type,
      })
      .then((response) => {
        return response.data;
      })
      .catch((error) => {
        console.log(error);
        toast.error("Ошибка при загрузке файла");
      });

    if (!data) {
      return;
    }

    const formData = new FormData();

    formData.append("file", file);

    const response = await fetch(data.uploadUrl!, {
      method: "PUT",
      body: file,
      headers: {
        "Content-Type": file.type,
      },
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error("Ошибка при загрузке файла");
        } else {
          toast.success("Фотография успешно загружена");
        }

        return response;
      })
      .catch((error) => {
        console.log(error);
        toast.error("Ошибка при загрузке файла");
      });

    if (!response) {
      return;
    }

    await authApi.v1
      .userServiceUpdateUser({
        profilePictureUrl: data.getUrl,
      })
      .catch((error) => {
        console.log(error);
        toast.error("Ошибка при обновлении аватара");
      });

    e.target.value = "";

    setProfilePictureURL(data.getUrl!);
  }

  return (
    <div className="flex flex-col gap-2 items-center justify-around pt-4 px-4">
      <Avatar src={src} />
      <label className="text-blue-500 cursor-pointer hover:underline text-sm">
        Загрузить фотографию
        <input
          accept="image/*"
          className="hidden"
          type="file"
          onChange={handleUploadAvatar}
        />
      </label>
    </div>
  );
}

export default function EditProfilePage() {
  const [user, setUser] = useState<WorkoutUser>();

  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);

  async function fetchMe() {
    await authApi.v1
      .userServiceGetMe()
      .then((response) => {
        console.log(response.data);
        setUser(response.data.user!);
      })
      .catch((error) => {
        console.log(error);
        throw error;
      });
  }

  async function fetchData() {
    setIsLoading(true);
    try {
      await fetchMe();
      setIsError(false);
    } catch {
      setIsError(true);
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    fetchData();
  }, []);

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return (
      <div className="p-4">
        <h2 className="text-lg text-red-500">Ошибка при загрузке данных</h2>
        <p>Проверьте соединение с сервером или обновите страницу.</p>
      </div>
    );
  }

  return (
    <div className="py-4 flex flex-col h-full">
      <PageHeader enableBackButton title="Редактировать профиль" />
      <AvatarSection
        setProfilePictureURL={(url) =>
          setUser({ ...user!, profilePictureUrl: url })
        }
        src={user!.profilePictureUrl!}
      />
      <DataForm user={user!} />
    </div>
  );
}
