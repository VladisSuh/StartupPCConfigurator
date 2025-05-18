import { useRef, useState } from 'react';
import './BrandsScroll.css'; // подключи стили, см. ниже

const BrandsScroll = ({ children }: { children: React.ReactNode }) => {
    const containerRef = useRef<HTMLDivElement>(null);
    const [isDown, setIsDown] = useState(false);
    const [startX, setStartX] = useState(0);
    const [scrollLeft, setScrollLeft] = useState(0);

    const handleMouseDown = (e: React.MouseEvent) => {
        setIsDown(true);
        const container = containerRef.current!;
        setStartX(e.pageX - container.offsetLeft);
        setScrollLeft(container.scrollLeft);
    };

    const handleMouseLeave = () => setIsDown(false);
    const handleMouseUp = () => setIsDown(false);

    const handleMouseMove = (e: React.MouseEvent) => {
        if (!isDown) return;
        e.preventDefault();
        const container = containerRef.current!;
        const x = e.pageX - container.offsetLeft;
        const walk = (x - startX) * 1.5; // чувствительность
        container.scrollLeft = scrollLeft - walk;
    };

    return (
        <div
            ref={containerRef}
            className="scroll-container"
            onMouseDown={handleMouseDown}
            onMouseLeave={handleMouseLeave}
            onMouseUp={handleMouseUp}
            onMouseMove={handleMouseMove}
        >
            {children}
        </div>
    );
};

export default BrandsScroll;
